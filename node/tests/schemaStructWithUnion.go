package tests

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/ipld/go-ipld-prime/fluent"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/schema"
)

func SchemaStructWithUnion(t *testing.T, engine Engine) {
	ts := schema.TypeSystem{}
	ts.Init()

	ts.Accumulate(schema.SpawnString("String"))

	// alpha is a struct that contains beta
	ts.Accumulate(schema.SpawnStruct("Alpha",
		[]schema.StructField{
			schema.SpawnStructField("beta", "Beta", false, false),
		},
		schema.SpawnStructRepresentationMap(map[string]string{}),
	))

	// beta is a union, contains gamma
	ts.Accumulate(schema.SpawnUnion("Beta",
		[]schema.TypeName{
			"Gamma",
		},
		schema.SpawnUnionRepresentationStringprefix(
			":",
			map[string]schema.TypeName{
				"gamma": "Gamma",
			},
		),
	))

	// gamma is a string
	ts.Accumulate(schema.SpawnString("Gamma"))
	engine.Init(t, ts)

	// ----------------------------------------
	// construct 3 nodes, 1 of each type

	gammaPrototype := engine.PrototypeByName("Gamma")
	gammaNode := fluent.MustBuild(gammaPrototype, func(na fluent.NodeAssembler) {
		na.AssignString("ok")
	})

	betaPrototype := engine.PrototypeByName("Beta")
	betaNode := fluent.MustBuildMap(betaPrototype, 1, func(ma fluent.MapAssembler) {
		ma.AssembleEntry("Gamma").AssignNode(gammaNode)
	})

	alphaPrototype := engine.PrototypeByName("Alpha")
	alphaNode := fluent.MustBuildMap(alphaPrototype, 1, func(ma fluent.MapAssembler) {
		ma.AssembleEntry("beta").AssignNode(betaNode)
	})

	// This works ok!
	res := printer.Sprint(alphaNode)
	expect := `struct<Alpha>{
	beta: union<Beta>{string<Gamma>{"ok"}}
}`
	qt.Check(t, res, qt.Equals, expect)

	// ----------------------------------------

	alphaReprPrototype := engine.PrototypeByName("Alpha.Repr")
	alphaReprNode := fluent.MustBuildMap(alphaReprPrototype, 1, func(ma fluent.MapAssembler) {
		// This panics:
		// panic: bindnode AssembleKey TODO: schema.UnionRepresentation_Stringprefix [recovered]
		//        panic: bindnode AssembleKey TODO: schema.UnionRepresentation_Stringprefix
		// It looks like since `Alpha.Repr` is using the representation, it tries
		// to assign to Beta (which is a Union) by treating it as a UnionRepr.
		// AssignNode calls `AssembleKey` on this UnionRepr, which currently panic's
		// since it is not implemented.
		ma.AssembleEntry("beta").AssignNode(betaNode)
	})

	res = printer.Sprint(alphaReprNode)
	expect = `struct<Alpha>{
	beta: union<Beta>{string<Gamma>{"ok"}}
}`
	qt.Check(t, res, qt.Equals, expect)
}
