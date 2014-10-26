package hub

import "testing"
import "bytes"

// empty primitives

func TestEmptyIsEmpty(t *testing.T) {
  empty := EmptyMessage()

  if (!empty.IsEmpty()) {
    t.Errorf("expecting: EmptyMessage to be empty")
  }
}

//non-empty primitives

func TestMessageIsEmpty(t *testing.T) {
  msg := NewMessage("sender", []byte("message"))

  if (msg.IsEmpty()) {
    t.Errorf("expecting: Message not to be empty")
  }
}

func TestMessageStr(t *testing.T) {
  msg := NewMessage("sender", []byte("message"))

  if (msg.Str() != "message") {
    t.Errorf("expecting: Message to have str")
  }
  if (bytes.Compare(msg.bytes, []byte("message")) != 0) {
    t.Errorf("expecting: Message to have bytes")
  }
}

// private, but protects other expectations
func TestMessageCount(t *testing.T) {
  msg := NewMessage("sender", []byte("message"))

  if (msg.count != 1) {
    t.Errorf("expecting: Message to have count")
  }
}

func TestMessageBytes1(t *testing.T) {
  msg := NewMessage("sender", []byte("message"))
  // msg.count = 1

  if (bytes.Compare(msg.Bytes(), []byte("message")) != 0) {
    t.Errorf("expecting: msg.count == 1")
  }
}

func TestMessageBytes5(t *testing.T) {
  msg := NewMessage("sender", []byte("message"))
  msg.count = 5

  if (bytes.Compare(msg.Bytes(), []byte("5:message")) != 0) {
    t.Errorf("expecting: msg.count == 1")
  }
}

// merge

func TestEmptyMergeEmpty(t *testing.T) {
  empty := EmptyMessage()
  empty2 := EmptyMessage()

  tgt := empty.Merge(empty2)

  if (!tgt.IsEmpty()) {
    t.Errorf("expecting: empty.merge(empty2) to be empty")
  }
}

func TestEmptyMergeMessage(t *testing.T) {
  empty := EmptyMessage()
  msg := NewMessage("sender", []byte("message"))
  tgt := empty.Merge(msg)

  // private
  if (tgt.sender != "sender") {
    t.Errorf("expecting: empty.merge(msg) to have sender")
  }

  if (bytes.Compare(tgt.Bytes(), []byte("message")) != 0) {
    t.Errorf("expecting: empty.merge(msg) to have bytes")
  }

  if (tgt.Str() != "message") {
    t.Errorf("expecting: empty.merge(msg) to have string")
  }
}

func TestMessageMergeMessage(t *testing.T) {
  msg := NewMessage("sender", []byte("message"))
  msg2 := NewMessage("sender", []byte("message"))

  tgt := msg.Merge(msg2)

  if (bytes.Compare(tgt.Bytes(), []byte("2:message")) != 0) {
    t.Errorf("expecting: msg.merge(msg) to have bytes")
  }

  if (tgt.Str() != "message") {
    t.Errorf("expecting: msg.merge(msg) to have string")
  }
}

func TestEmptyMergeMessageMergeMessage(t *testing.T) {
  empty := EmptyMessage()
  msg := NewMessage("sender", []byte("message"))
  tmp := empty.Merge(msg)
  tgt := tmp.Merge(msg)

  // private
  if (msg.sender != "sender") {
    t.Errorf("expecting: empty.merge(msg) to have sender")
  }

  if (bytes.Compare(tgt.Bytes(), []byte("2:message")) != 0) {
    t.Errorf("expecting: empty.merge(msg) to have bytes")
  }

  if (tgt.Str() != "message") {
    t.Errorf("expecting: empty.merge(msg) to have string")
  }
}

