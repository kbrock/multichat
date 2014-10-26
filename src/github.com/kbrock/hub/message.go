package hub

import "strconv"

type message struct {
  sender string
  count int
  bytes []byte
}

func (m *message) merge(src *message) {
  if src.count == 0 {
    m.count += 1
  } else {
    m.count += src.count
  }
}

func (m *message) Merge2(src *message) *message {
  if m.IsEmpty() {
    return &message{sender:src.sender, bytes: src.bytes, count: src.count}
  } else {
    return &message{sender:m.sender, bytes: m.bytes, count: m.count + src.count}
  }
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

func (m *message) allBytes() []byte {
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

// TODO: Str
func (m *message) str() string {
  return string(m.bytes)
}

// going away
func (m *message) clone() *message {
  return &message{sender: m.sender, count: 1, bytes: m.bytes};
}

func (m *message) IsEmpty() bool {
  return m.count == 0
}
