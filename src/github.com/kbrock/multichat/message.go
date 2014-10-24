package main

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

func (m *message) str() string {
  return string(m.bytes)
}

func (m *message) clone() *message {
  return &message{sender: m.sender, count: 1, bytes: m.bytes};
}
