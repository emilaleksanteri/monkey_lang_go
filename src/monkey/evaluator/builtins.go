package evaluator

import (
	"fmt"
	"monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
	},
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	},
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument 'first' mut be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		},
	},
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument 'last' mut be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)

			// more efficient to do length before we use it twice in case array is really big
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return NULL
		},
	},
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument 'rest' mut be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElems := make([]object.Object, length-1, length-1)
				copy(newElems, arr.Elements[1:length])
				return &object.Array{Elements: newElems}
			}

			return NULL

		},
	},
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument 'push' mut be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			newElems := make([]object.Object, length+1, length+1)
			copy(newElems, arr.Elements)
			newElems[length] = args[1]

			return &object.Array{Elements: newElems}
		},
	},
	"pop": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument 1 'pop' mut be ARRAY, got %s", args[0].Type())
			}

			if args[1].Type() != object.INTEGER_OBJ {
				return newError("argument 2 'pop' mut be INTEGER, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			idx := args[1].(*object.Integer)

			length := len(arr.Elements)
			if length > 0 && (int(idx.Value) < length && int(idx.Value) >= 0) {
				newElems := []object.Object{}
				for elemIdx, elem := range arr.Elements {
					if elemIdx == int(idx.Value) {
						continue
					}

					newElems = append(newElems, elem)
				}

				return &object.Array{Elements: newElems}
			}

			if int(idx.Value) >= length || int(idx.Value) < 0 {
				return newError("index out of range")
			}

			return NULL

		},
	},
}
