package heap

import "fmt"
import "JVM/classfile"
import "JVM/classpath"

/*
class names:
    - primitive types: boolean, byte, int ...
    - primitive arrays: [Z, [B, [I ...
    - non-array classes: java/lang/Object ...
    - array classes: [Ljava/lang/Object; ...
*/
type ClassLoader struct {
	cp               *classpath.Classpath
	classMap         map[string]*Class // loaded classes/
	verboseClassFlag bool
}

func NewClassLoader(cp *classpath.Classpath, verboseClassFlag bool) *ClassLoader {
	loader := &ClassLoader{
		cp:               cp,
		verboseClassFlag: verboseClassFlag,
		classMap:         make(map[string]*Class),
	}
	loader.loadBasicClasses()
	loader.loadPrimitiveClasses()
	return loader
}

func (loader *ClassLoader) LoadClass(name string) *Class {
	if class, ok := loader.classMap[name]; ok {
		// already loaded
		return class
	}

	var class *Class
	if name[0] == '[' {
		class = loader.loadArrayClass(name, loader.verboseClassFlag)
	} else {
		class = loader.loadNonArrayClass(name, loader.verboseClassFlag)
	}

	if jlClassClass, ok := loader.classMap["java/lang/Class"]; ok {
		class.jClass = jlClassClass.NewObject()
		class.jClass.extra = class
	}

	return class
}

func (loader *ClassLoader) loadPrimitiveClasses() {
	for primitiveType, _ := range primitiveTypes {
		loader.loadPrimitiveClass(primitiveType)
	}
}

func (loader *ClassLoader) loadPrimitiveClass(name string) {
	class := &Class{
		accessFlags: ACC_PUBLIC,
		name:        name,
		loader:      loader,
		inited:      true,
	}
	class.jClass = loader.classMap["java/lang/Class"].NewObject()
	class.jClass.extra = class
	loader.classMap[name] = class
}

func (loader *ClassLoader) loadBasicClasses() {
	classClass := loader.LoadClass("java/lang/Class")
	for _, class := range loader.classMap {
		if class.jClass == nil {
			class.jClass = classClass.NewObject()
			class.jClass.extra = class
		}
	}
}

func (loader *ClassLoader) loadNonArrayClass(name string, verboseClassFlag bool) *Class {
	data, entry := loader.readClass(name)
	class := loader.defineClass(data)
	link(class)

	if verboseClassFlag {
		fmt.Printf("[Loaded %s from %s]\n", name, entry)
	}
	return class
}

func (loader *ClassLoader) loadArrayClass(name string, verboseClassFlag bool) *Class {
	class := &Class{
		accessFlags: ACC_PUBLIC,
		name:        name,
		loader:      loader,
		inited:      true,
		superClass:  loader.LoadClass("java/lang/Object"),
		interfaces: []*Class{
			loader.LoadClass("java/lang/Cloneable"),
			loader.LoadClass("java/io/Serializable"),
		},
	}
	loader.classMap[name] = class

	if verboseClassFlag {
		fmt.Printf("[Loaded Array Class %s]\n", name)
	}
	return class
}

func (loader *ClassLoader) readClass(name string) ([]byte, classpath.Entry) {
	data, entry, err := loader.cp.ReadClass(name)
	if err != nil {
		panic("java.lang.ClassNotFoundException: " + name)
	}
	return data, entry
}

// jvm spec 5.3.5
func (loader *ClassLoader) defineClass(data []byte) *Class {
	class := parseClass(data)
	class.loader = loader
	resolveSuperClass(class)
	resolveInterfaces(class)
	loader.classMap[class.name] = class
	return class
}

func parseClass(data []byte) *Class {
	cf, err := classfile.Parse(data)
	if err != nil {
		//panic("java.lang.ClassFormatError")
		panic(err)
	}
	return newClass(cf)
}

// jvm spec 5.4.3.1
// recursive call until Object.class, load -> define -> superclass(load again)
func resolveSuperClass(class *Class) {
	if class.name != "java/lang/Object" {
		class.superClass = class.loader.LoadClass(class.superClassName)
	}
}
func resolveInterfaces(class *Class) {
	interfaceCount := len(class.interfaceNames)
	if interfaceCount > 0 {
		class.interfaces = make([]*Class, interfaceCount)
		for i, interfaceName := range class.interfaceNames {
			class.interfaces[i] = class.loader.LoadClass(interfaceName)
		}
	}
}

func link(class *Class) {
	verify(class)
	prepare(class)
}

func verify(class *Class) {
	// todo
}

// jvm spec 5.4.2
func prepare(class *Class) {
	calcInstanceFieldSlotIds(class)
	calcStaticFieldSlotIds(class)
	allocAndInitStaticVars(class)
}

func calcInstanceFieldSlotIds(class *Class) {
	slotId := uint(0)
	if class.superClass != nil {
		slotId = class.superClass.instanceSlotCount
	}
	for _, field := range class.fields {
		if !field.IsStatic() {
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() {
				slotId++
			}
		}
	}
	class.instanceSlotCount = slotId
}

func calcStaticFieldSlotIds(class *Class) {
	slotId := uint(0)
	for _, field := range class.fields {
		if field.IsStatic() {
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() {
				slotId++
			}
		}
	}
	class.staticSlotCount = slotId
}

func allocAndInitStaticVars(class *Class) {
	class.staticVars = newSlots(class.staticSlotCount)
	for _, field := range class.fields {
		if field.IsStatic() && field.IsFinal() {
			initStaticFinalVar(class, field)
		}
	}
}

func initStaticFinalVar(class *Class, field *Field) {
	vars := class.staticVars
	cp := class.constantPool
	cpIndex := field.ConstValueIndex()
	slotId := field.SlotId()

	if cpIndex > 0 {
		switch field.Descriptor() {
		case "Z", "B", "C", "S", "I":
			val := cp.GetConstant(cpIndex).(int32)
			vars.SetInt(slotId, val)
		case "J":
			val := cp.GetConstant(cpIndex).(int64)
			vars.SetLong(slotId, val)
		case "F":
			val := cp.GetConstant(cpIndex).(float32)
			vars.SetFloat(slotId, val)
		case "D":
			val := cp.GetConstant(cpIndex).(float64)
			vars.SetDouble(slotId, val)
		case "Ljava/lang/String;":
			goStr := cp.GetConstant(cpIndex).(string)
			jStr := JString(class.Loader(), goStr)
			vars.SetRef(slotId, jStr)
		}
	}
}
