package reachability

import objecttype "lindenii.org/go/furgit/object/type"

func (walk *Walk) initialStack() []walkItem {
	if len(walk.wants) == 0 {
		return nil
	}

	stack := make([]walkItem, 0, len(walk.wants))
	for want := range walk.wants {
		stack = append(stack, walkItem{id: want, want: objecttype.TypeInvalid})
	}

	return stack
}
