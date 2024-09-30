package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	// Place your code here
	Head   *ListItem
	Last   *ListItem
	Length int
}

func (l *list) Len() int {
	return l.Length
}

func (l *list) Front() *ListItem {
	return l.Head
}

func (l *list) Back() *ListItem {
	return l.Last
}

func (l *list) PushFront(v interface{}) *ListItem {
	if v != nil {
		newItem := &ListItem{Value: v}
		if l.Length == 0 {
			l.Head = newItem
			l.Last = newItem
		} else {
			newItem.Next = l.Head
			l.Head.Prev = newItem
			l.Head = newItem
		}
		l.Length++
		return newItem
	}
	return nil
}

func (l *list) PushBack(v interface{}) *ListItem {
	if v != nil {
		newItem := &ListItem{Value: v}
		if l.Length == 0 {
			l.Head = newItem
			l.Last = newItem
		} else {
			newItem.Prev = l.Last
			l.Last.Next = newItem
			l.Last = newItem
		}
		l.Length++
		return newItem
	}
	return nil
}

func (l *list) Remove(i *ListItem) {
	if i == nil || l.Length == 0 {
		return
	}
	if i == l.Head {
		l.Head = i.Next
	} else {
		i.Prev.Next = i.Next
	}
	if i == l.Last {
		l.Last = i.Prev
	} else {
		i.Next.Prev = i.Prev
	}

	i.Prev = nil
	i.Next = nil
	l.Length--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || l.Length == 0 {
		return
	}
	if i == l.Head {
		return
	}
	l.Remove(i)

	i.Next = l.Head
	l.Head.Prev = i

	l.Head = i
	i.Prev = nil

	l.Length++
}

func NewList() List {
	return new(list)
}
