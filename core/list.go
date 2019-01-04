
// List data structure

package golsp

type Item struct {
	next *Item
	Object Object
}

type List struct {
	First *Item
	Last *Item
	Length int
	branches map[int]*Item
}

func ListFromSlice(slice []Object) List {
	l := List{}
	for _, obj := range slice { l.Append(obj) }
	return l
}

func (l *List) at(index int) *Item {
	if index < 0 || index >= l.Length { return nil }
	if index == l.Length - 1 { return l.Last }

	current := l.First
	for i := 0; i < index; i++ { current = l.Next(current, i) }

	return current
}

func (l *List) slice(begin int, end int) List {
	if end <= begin { return List{} }

	slice := List{
		First: l.at(begin),
		Last: l.at(end - 1),
		Length: end - begin,
		branches: make(map[int]*Item),
	}

	for index, branch := range l.branches {
		if index >= begin && index < end {
			slice.branches[index - begin] = branch
		}
	}

	return slice
}

func (l *List) sublist(begin *Item, index int) List {
	sublist := List{
		First: begin,
		Last: l.Last,
		Length: l.Length - index,
		branches: make(map[int]*Item),
	}

	for i, branch := range l.branches {
		if i >= index { sublist.branches[i] = branch }
	}

	return sublist
}

func (l *List) Next(item *Item, index int) *Item {
	if (index >= l.Length) { return nil }

	branch, exists := l.branches[index]
	if exists { return branch }

	return item.next
}

func (l *List) ToSlice() []Object {
	slice := make([]Object, 0, l.Length)
	ptr := l.First
	for i := 0; i < l.Length; i++ {
		slice = append(slice, ptr.Object)
		ptr = l.Next(ptr, i)
	}

	return slice
}

func (l *List) Copy() List {
	copy := List{}
	slice := l.ToSlice()
	for _, obj := range slice { copy.Append(CopyObject(obj)) }

	return copy
}

func (l *List) Append(obj Object) {
	newitem := Item{Object: obj}
	defer func() {
		l.Length++
		l.Last = &newitem
	}()

	if l.Length == 0 {
		l.First = &newitem
		return
	}

	if l.Last.next == nil {
		l.Last.next = &newitem
		return
	}

	if l.branches == nil { l.branches = make(map[int]*Item) }
	l.branches[l.Length - 1] = &newitem
}

func (self *List) Join(other List) {
	defer func() {
		for index, branch := range other.branches {
			self.branches[index + self.Length] = branch
		}
		self.Length += other.Length
		self.Last = other.Last
	}()

	if self.Length == 0 {
		self.First = other.First
		return
	}

	if self.Last.next == nil {
		self.Last.next = other.First
		return
	}

	if self.branches == nil { self.branches = make(map[int]*Item) }
	self.branches[self.Length - 1] = other.First
}

func (l *List) Index(index int) Object {
	if index < 0 { index += l.Length }

	item := l.at(index)
	if item == nil { return UndefinedObject() }
	return item.Object
}

func (l *List) Slice(begin int, end int) Object {
	if begin < 0 { begin += l.Length }
	if begin < 0 || begin >= l.Length { return UndefinedObject() }
	if end < 0 { end += l.Length }
	if end < 0 { return UndefinedObject() }
	if end > l.Length { end = l.Length }

	return Object{
		Type: ObjectTypeList,
		Elements: l.slice(begin, end),
	}
}

func (l *List) SliceStep(begin int, end int, step int, sliceAll bool) Object {
	if begin < 0 { begin += l.Length }
	if begin < 0 || begin >= l.Length { return UndefinedObject() }
	if end < 0 && !sliceAll { end += l.Length }
	if end < 0 && !sliceAll { return UndefinedObject() }
	if end > l.Length { end = l.Length }

	slice := l.ToSlice()
	newlist := List{}
	for i := begin; i != end; i += step {
		if i >= len(slice) { break }
		if i < 0 { break }

		newlist.Append(slice[i])
	}

	return Object{
		Type: ObjectTypeList,
		Elements: newlist,
	}
}
