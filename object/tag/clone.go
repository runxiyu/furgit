package tag

import "bytes"

// Clone returns a deep copy of the tag
// whose byte fields are independent of any memory the original may alias.
//
// Labels: Life-Independent.
func (tag *Tag) Clone() *Tag {
	clone := &Tag{
		TargetID:   tag.TargetID,
		TargetType: tag.TargetType,
		Name:       bytes.Clone(tag.Name),
		Tagger:     tag.Tagger.Clone(),
		Message:    bytes.Clone(tag.Message),
	}

	if tag.ExtraHeaders != nil {
		clone.ExtraHeaders = make([]ExtraHeader, len(tag.ExtraHeaders))
		for i, h := range tag.ExtraHeaders {
			clone.ExtraHeaders[i] = ExtraHeader{
				Key:   bytes.Clone(h.Key),
				Value: bytes.Clone(h.Value),
			}
		}
	}

	return clone
}
