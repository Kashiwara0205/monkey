package evaluator

import(
	"../lexer"
	"../object"
	"../parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T){
	tests := []struct{
		input string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests{

		//Evalは１ぽんずっこ
		evaluated := testEval(tt.input)

		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object{
	// 字句解析
	l := lexer.New(input)
	// 構文解析
	p := parser.New(l)
	program := p.ParseProgram()
	// Evalには初期にprogramノード(全ノードのトップ)を渡す
	env := object.NewEnviroment()
	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	// astの時にもあった変換手法
	result, ok := obj.(*object.Integer)

	// 変換できたかどうか	
	if !ok{
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}

	// expectedは数値
	// 数値がastから代入されてオブジェクトとして帰ってくるか
	if result.Value != expected{
		t.Errorf("Object has wrong value. got=%d, want=%d",
			result.Value, expected)
	}

	return true
}

func TestEvalBooleanExpression(t *testing.T){
	tests := []struct{
		input string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 !=2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests{
		// 評価終了
		evaluated := testEval(tt.input)

		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool{
	result, ok := obj.(*object.Boolean)
	if !ok{
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	// 実際にastから真偽値が格納されたか
	if result.Value != expected{
		t.Errorf("object has wrong value. got = %t, want = %t",
			result.Value, expected)
	}

	return true
}

func TestBangOperator(t *testing.T){
	tests := []struct{
		input string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests{
		// 評価完了
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T){
	tests := []struct{
		input string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests{
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok{
			testIntegerObject(t, evaluated, int64(integer))
		}else{
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) bool{
	if obj != NULL{
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func TestReturnStatements(t *testing.T){
	tests := []struct{
		input string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 *5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{
			`
			if (10 > 1){
				if (10 > 1){
					return 10;
				}
				return 1;
			}
			`,
			10,
		},
	}

	for _, tt := range tests{
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T){
	tests := []struct{
		input string
		expectedMessage string
	}{
		{
			"5 + true;", 
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
			if (10 > 1){
				if (10 > 1){
					return true + false;
				}
				return 1;
			}
			`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not foutnd:foobar",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
	}

	for _, tt := range tests{
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok{
			t.Errorf("no error objct returned. got=%T(%+v)",
					 evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage{
			t.Errorf("wrong error message. expected=%q, got=%q",
					tt.expectedMessage, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T){
	tests := []struct{
		input string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests{
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject (t *testing.T){
	input := "fn(x) { x + 2; };"

	evaluted := testEval(input)

	// functionオブジェクトに変換
	fn, ok := evaluted.(*object.Function)

	// 変換ができたか
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluted, evaluted)
	}

	// パラメータの数
	if len(fn.Parameters) != 1{
		t.Fatalf("Function has wrong parameters. Prameters = %+v",
			fn.Parameters)
	}

	// パラメータの内容
	if fn.Parameters[0].String() != "x"{
		t.Fatalf("parameter is not 'x'. got =%q", fn.Parameters[0])
	}

	// 関数の中身
	expectedBody := "(x + 2)"
	if fn.Body.String()!= expectedBody{
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T){
	// 関数のcallに関するテスト
	tests := []struct{
		input string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests{
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestStringLiteral(t *testing.T){
	input := `"Hello world"`

	evaluted := testEval(input)
	str, ok := evaluted.(*object.String)
	if !ok{
		t.Fatalf("object is not String. got=%T (%+v)", evaluted, evaluted)
	}

	if str.Value != "Hello world"{
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T){
	input := `"Hello" + "World!"`

	evaluted := testEval(input)
	str, ok := evaluted.(*object.String)
		if !ok{
			t.Fatalf("object is not String. got=%T (%+v)", evaluted, evaluted)
		}

		if str.Value != "HelloWorld!"{
			t.Errorf("String has wrong value. got=%q", str.Value)
		}
}

func TestBuiltinFunctions(t *testing.T){
	tests := []struct{
		input string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tests{
		evaluted := testEval(tt.input)

		switch expected := tt.expected.(type){
		case int:
			testIntegerObject(t, evaluted, int64(expected))
		case string:
			errObj, ok := evaluted.(*object.Error)
			if !ok{
				t.Errorf("object is not Error. got = %T (%+v)",
					evaluted, evaluted)

				continue
			}
			if errObj.Message != expected{
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		}
	}
}

func TestArrayLiterals(t *testing.T){
	input := "[1, 2 * 2, 3 + 3]"

	evaluted := testEval(input)
	result, ok := evaluted.(*object.Array)
	if !ok{
		t.Fatalf("object is not Array . got=%T (%+v)", evaluted, evaluted)
	}

	if len(result.Elements) != 3{
		t.Fatalf("array has wrong num of elements. got=%d",
			len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T){
	tests := []struct{
		input string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"let i = 0;[1][i]",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,

		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}

	for _, tt := range tests{
		evaluted := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok{
			// 上記構造体と、配列の出力結果を検証
			testIntegerObject(t, evaluted, int64(integer))
		}else{
			// おかしな場所を指し示していたのならnullを表示させる
			testNullObject(t, evaluted)
		}
	}
}