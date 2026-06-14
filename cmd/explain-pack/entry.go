package main

import (
	"bytes"
	"fmt"
	"io"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/format/packfile/delta"
	"lindenii.org/go/furgit/object/tree"
	"lindenii.org/go/lgo/intconv"
)

func (explainer *explainer) explainEntry(num, count, cursor int) (int, error) {
	hashSize := explainer.objectFormat.Size()

	header, err := packfile.ParseEntryHeader(explainer.data[cursor:], hashSize)
	if err != nil {
		return 0, fmt.Errorf("entry %d at offset %d: %w", num, cursor, err)
	}

	payloadStart := cursor + header.HeaderLen
	if payloadStart > len(explainer.data) {
		return 0, fmt.Errorf("entry %d at offset %d: header runs past the end of the pack", num, cursor)
	}

	payload, consumed, err := inflateAt(explainer.data[payloadStart:])
	if err != nil {
		return 0, fmt.Errorf("entry %d at offset %d: %w", num, cursor, err)
	}

	next := payloadStart + consumed

	explainer.printf("object %d of %d\n", num, count)
	explainer.printf("\tty\t%s\n", entryTypeLabel(header.Type))
	explainer.printf("\tofs\t%d\n", cursor)
	explainer.printf("\thdrsz\t%d\n", header.HeaderLen)
	explainer.printf("\tsz\t%d\n", header.Size)

	if uint64(len(payload)) != header.Size {
		explainer.printf("\tnote\tdeclared %d byte(s) but inflated to %d\n", header.Size, len(payload))
	}

	if header.Type.IsBase() {
		err = explainer.renderBase(cursor, header.Type, payload, consumed)
	} else {
		err = explainer.renderDelta(cursor, header, payload, consumed)
	}

	if err != nil {
		return 0, fmt.Errorf("entry %d at offset %d: %w", num, cursor, err)
	}

	explainer.printf("\n")

	return next, nil
}

func (explainer *explainer) renderBase(cursor int, entryType packfile.EntryType, content []byte, consumed int) error {
	explainer.renderContent(entryType, content)

	explainer.printf("\tzlib\t%d\n", consumed)

	oid, err := explainer.recomputeOID(entryType, content)
	if err != nil {
		return err
	}

	explainer.printf("\toid\t%s\n", oid)

	explainer.oidIndex[oid] = cursor
	explainer.cache.Add(cursor, resolvedBase{entryType: entryType, content: content})

	return nil
}

func (explainer *explainer) renderDelta(cursor int, header packfile.EntryHeader, payload []byte, consumed int) error {
	baseSize, resultSize, pos, err := delta.ParseHeaderSizes(payload)
	if err != nil {
		return fmt.Errorf("delta header: %w", err)
	}

	err = explainer.renderBaseRef(cursor, header)
	if err != nil {
		return err
	}

	explainer.printf("\tbasesz\t%d\n", baseSize)
	explainer.printf("\tnewsz\t%d\n", resultSize)

	baseOffset, located, err := explainer.baseOffset(cursor, header)
	if err != nil {
		return err
	}

	var (
		baseType     packfile.EntryType
		baseContent  []byte
		baseResolved bool
	)

	if located {
		baseType, baseContent, baseResolved, err = explainer.reconstruct(baseOffset, 0)
		if err != nil {
			return err
		}
	}

	var walkBase []byte
	if baseResolved {
		walkBase = baseContent
	}

	result, complete, err := explainer.walkDelta(walkBase, payload, pos)
	if err != nil {
		return err
	}

	explainer.printf("\tzlib\t%d\n", consumed)

	switch {
	case baseResolved && complete:
		if uint64(len(result)) != resultSize {
			explainer.printf("\tnote\tdelta produced %d byte(s) but declared %d\n", len(result), resultSize)
		}

		explainer.renderContent(baseType, result)

		newOID, err := explainer.recomputeOID(baseType, result)
		if err != nil {
			return err
		}

		explainer.printf("\tnewoid\t%s\n", newOID)

		explainer.oidIndex[newOID] = cursor
		explainer.cache.Add(cursor, resolvedBase{entryType: baseType, content: result})
	case !baseResolved:
		explainer.printf("\tnote\tbase not available in this pack; cannot reconstruct\n")
	default:
		explainer.printf("\tnote\tdelta decode incomplete; cannot reconstruct\n")
	}

	return nil
}

func (explainer *explainer) renderBaseRef(cursor int, header packfile.EntryHeader) error {
	switch header.Type {
	case packfile.EntryTypeOfsDelta:
		dist, err := intconv.Uint64ToInt(header.OfsDistance)
		if err != nil {
			return fmt.Errorf("ofs-delta distance overflows int: %w", err)
		}

		explainer.printf("\tbaseofs\t-%d = %d\n", dist, cursor-dist)
	case packfile.EntryTypeRefDelta:
		baseID, err := explainer.objectFormat.FromBytes(header.RefBase[:explainer.objectFormat.Size()])
		if err != nil {
			return fmt.Errorf("ref-delta base ID: %w", err)
		}

		explainer.printf("\tbaseoid\t%s\n", baseID)
	case packfile.EntryTypeInvalid,
		packfile.EntryTypeCommit,
		packfile.EntryTypeTree,
		packfile.EntryTypeBlob,
		packfile.EntryTypeTag,
		packfile.EntryTypeFuture:
	}

	return nil
}

func (explainer *explainer) renderContent(entryType packfile.EntryType, content []byte) {
	switch entryType {
	case packfile.EntryTypeCommit, packfile.EntryTypeTag:
		explainer.printf("\tcontent\n")
		indentBlock(explainer.out, "\t\t", content)
	case packfile.EntryTypeTree:
		explainer.renderTree(content)
	case packfile.EntryTypeBlob,
		packfile.EntryTypeOfsDelta,
		packfile.EntryTypeRefDelta,
		packfile.EntryTypeInvalid,
		packfile.EntryTypeFuture:
		explainer.printf("\thexdump\n")
		hexBlock(explainer.out, "\t\t", content)
	}
}

func (explainer *explainer) renderTree(content []byte) {
	parsed, err := tree.Parse(content, explainer.objectFormat)
	if err != nil {
		explainer.printf("\thexdump\t(not a valid tree: %v)\n", err)
		hexBlock(explainer.out, "\t\t", content)

		return
	}

	explainer.printf("\ttree\n")

	for _, entry := range parsed.Entries() {
		mode := string(entry.Mode.Append(nil))
		explainer.printf(
			"\t\t%s %s %s\t%s\n",
			mode, entry.Mode.ObjectType().Name(), entry.ID, entry.Name,
		)
	}
}

func inflateAt(data []byte) ([]byte, int, error) {
	reader := bytes.NewReader(data)

	zr, err := zlib.NewReader(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("opening zlib stream: %w", err)
	}

	content, err := io.ReadAll(zr)
	closeErr := zr.Close()

	if err != nil {
		return nil, 0, fmt.Errorf("inflating payload: %w", err)
	}

	if closeErr != nil {
		return nil, 0, fmt.Errorf("closing zlib stream: %w", closeErr)
	}

	consumed := len(data) - reader.Len()

	return content, consumed, nil
}

func entryTypeLabel(entryType packfile.EntryType) string {
	switch entryType {
	case packfile.EntryTypeCommit:
		return "commit"
	case packfile.EntryTypeTree:
		return "tree"
	case packfile.EntryTypeBlob:
		return "blob"
	case packfile.EntryTypeTag:
		return "tag"
	case packfile.EntryTypeOfsDelta:
		return "ofs-delta"
	case packfile.EntryTypeRefDelta:
		return "ref-delta"
	case packfile.EntryTypeInvalid:
		return "invalid"
	case packfile.EntryTypeFuture:
		return "future"
	default:
		return fmt.Sprintf("unknown (%d)", entryType)
	}
}
