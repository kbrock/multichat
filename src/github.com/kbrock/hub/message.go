package hub

import "strconv"

type message struct {
  sender string
  count int
  bytes []byte
}

func EmptyMessage() *message {
  return &message{}
}

func NewMessage(sender string, bytes []byte) *message{
  return &message{
    sender: sender,
    bytes: bytes,
    count: 1,
  }
}

func (m *message) Merge(src *message) *message {
  if m.IsEmpty() {
    return &message{sender:src.sender, bytes: src.bytes, count: src.count}
  } else {
    return &message{sender:m.sender, bytes: m.bytes, count: m.count + src.count}
  }
}

func (m *message) Bytes() []byte {
  if m.count <= 1 {
    return m.bytes
  } else {
    countS := strconv.Itoa(m.count)
    target := make([]byte, 0, 1 + len(countS) + len(m.bytes))
    target  = append(target, countS...)
    target  = append(target, ':')
    target  = append(target, m.bytes...)
    return target
  }
}

func (m *message) Str() string {
  return string(m.bytes)
}


func (m *message) IsEmpty() bool {
  return m.count == 0
}
