// Credit to steve wang, software engineer, zhuhai city, China for this stack implementation.
//
// https://groups.google.com/forum/#!topic/golang-nuts/xYJUfbIPrD0
//
// EveryBitCounts Software Services Inc. (EGH) added PopIf method.

// this package provides a thread-safe stack of arbitrary capacity.
// It has a non-blocking Push method,
// A blocking Pop method that blocks if the stack is empty,
// and a non-blocking PopIf method that returns nil if the stack is empty.

package thread_safe_stack

type Stack struct {
  in chan interface{}
  out chan interface{}
}

func NewStack() *Stack {
  p := new(Stack)
  p.in = make(chan interface{})
  p.out = make(chan interface{})
  go p.run()
  return p
}

func (p *Stack) run() {
  var objs []interface{}
  var out chan interface{}
  var top interface{}
  for {
    select {
    case v := <-p.in:
      if top != nil {
        objs = append(objs, top)
      }
      top = v
      out = p.out
    case out <- top:
      if len(objs) > 0 {
        top = objs[len(objs)-1]
        objs = objs[:len(objs)-1]
      } else {
        top = nil
        out = nil
      }
    }
  }
}

/*
Push an object onto the top of the stack.
Does not block.
*/
func (p *Stack) Push(d interface{}) {
  p.in <- d
}

/*
Returns the top value from the stack. Blocks until there is
a value in the stack.
*/
func (p *Stack) Pop() interface{} {
  return <-p.out
}

/*
Returns the top value from the stack, or nil if the stack is empty. 
Does not block. Added by EGH to steve wang stack implementation. 
Use freely.
*/
func (p *Stack) PopIf() (val interface{}) {
  select {
     case val = <-p.out:
     default: 
  }
  return
}
