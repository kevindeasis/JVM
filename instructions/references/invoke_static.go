package references

import (
	"JVM/instructions/base"
	"JVM/rtdz"
	"JVM/rtdz/heap"
)

// invoke a static method
type INVOKE_STATIC struct {
	base.Index16Instruction
}

func (invoke *INVOKE_STATIC) Execute(frame *rtdz.Frame) {
	pool := frame.Method().Class().ConstantPool()
	methodRef := pool.GetConstant(invoke.Index).(*heap.MethodRef)
	resolvedMethod := methodRef.ResolvedMethod()
	class := resolvedMethod.Class()

	if !class.Inited() {
		frame.RevertNextPC()
		base.InitClass(frame.Thread(), class)
		return
	}

	if !resolvedMethod.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}

	base.InvokeMethod(frame, resolvedMethod)
}
