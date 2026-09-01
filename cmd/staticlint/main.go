// Package main реализует статический анализатор кода staticlint.
//
// Сборка и запуск из корня проекта:
//
//	go build -o staticlint ./cmd/staticlint/
//
//	./staticlint ./...
//
// # Назначение подключаемых анализаторов
//
// Анализаторы пакета go/analysis/passes:
//   - appends выявляет некорректные вызовы append без присвоения результата в слайс.
//   - asmdecl проверяет соответствие сигнатур Go-функций их ассемблерным реализаций.
//   - assign обнаруживает присваивания переменных самим себе.
//   - atomic контролирует правильное использование атомарных функций пакета sync/atomic.
//   - bools выявляет синтаксические или логические ошибки в булевых операторах.
//   - buildtag проверяет корректность оформления директив условной сборки.
//   - cgocall выявляет нарушения правил передачи указателей памяти Go в функции C-кода через Cgo.
//   - composite контролирует обязательную инициализацию структур по именам полей.
//   - copylock выявляет копирование структур, содержащих Mutex или WaitGroup.
//   - defers выявляет логические ошибки в выражениях defer.
//   - directive проверяет корректность написания служебных Go-комментариев компилятора.
//   - errorsas проверяет, что в errors.As вторым параметром передан валидный указатель на ошибку.
//   - framepointer выявляет ошибки манипуляций с указателем фрейма стека в ассемблерном коде.
//   - hostport проверяет синтаксическую корректность сетевых адресов.
//   - httpcall выявляет пропущенные закрытия дескрипторов HTTP-ответов.
//   - ifaceassert выявляет невозможные утверждения интерфейсных типов.
//   - loopclosure выявляет баги захвата переменных цикла внутри горутин.
//   - lostcancel контролирует обязательный вызов возвращаемой функции отмены (cancel) для контекстов.
//   - nilfunc выявляет ошибочные и избыточные сравнения с nil.
//   - printf проверяет валидность строк форматирования данных.
//   - shift выявляет ошибки в операциях битового сдвига.
//   - sigchanyzer проверяет правильность емкости каналов для прослушивания сигналов пакета os/signal.
//   - slog проверяет вызовы структурного логирования log/slog на валидность пар ключ-значение.
//   - stdmethods проверяет сигнатуры методов на точное совпадение со стандартными интерфейсами.
//   - stdversion проверяет совместимость кода с версией языка Go, прописанной в go.mod.
//   - stringintconv выявляет потенциально ошибочные преобразования числовых типов в строки.
//   - structtag проверяет синтаксическую корректность разметки полей структур.
//   - testinggoroutine выявляет вызовы t.Fatal внутри горутин в тестах.
//   - tests проверяет правильность названий, сигнатур и расположения функций тестирования.
//   - timeformat выявляет ошибочные макеты в методах time.Format и time.Parse.
//   - unmarshal выявляет передачу не-указателей в функции json.Unmarshal.
//   - unreachable выявляет недостижимые блоки кода, идущие после выражений return/panic.
//   - unsafeptr выявляет ошибочные и некорректные преобразования выражений uintptr обратно в unsafe.Pointer.
//   - unusedresultsult выявляет скрытые логические ошибки, при которых результат вызова функций был проигнорирован.
//   - waitgroup выявляет ошибочные вызовы методов Add и Done у sync.WaitGroup.
//
// Анализаторы пакета staticcheck.io:
//   - класс SA набор базовых проверок для диагностики скрытых логических ошибок, выявления утечек ресурсов и использования устаревшего кода.
//   - S1002 упрощает избыточные конструкции булевых условий.
//   - ST1003 выявляет нарушения официального кодстайла Go при написании идентификаторов.
//   - QF1003 выявляет цепочки if/else-if для их преобразования в компактный переключатель switch.
//
// Внешние анализаторы:
//   - errcheck выявляет вызовы функций и методов, результаты ошибок которых были проигнорированы.
//   - bodyclose выявляет пропущенные вызовы закрытия дескрипторов HTTP-соединений.
//
// Кастомный анализатор:
//   - osExitCheck выявляет вызовы функции os.Exit внутри функции main пакета main.
package main

import (
	"go/ast"
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/hostport"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

var osExitCheckAnalyzer = &analysis.Analyzer{
	Name: "osExitCheck",
	Doc:  "check for os.Exit calls in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			decl, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}
			name := pass.Fset.Position(decl.Pos()).Filename
			if strings.Contains(name, "go-build") {
				return true
			}
			if decl.Name.Name == "main" {
				ast.Inspect(decl.Body, func(innerNode ast.Node) bool {
					if call, ok := innerNode.(*ast.CallExpr); ok {
						if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := selector.X.(*ast.Ident); ok {
								if ident.Name == "os" && selector.Sel.Name == "Exit" {
									pass.Reportf(call.Pos(), "обнаружен вызов os.Exit")
								}
							}
						}
					}
					return true
				})
				return false
			}
			return true
		})
	}
	return nil, nil
}

func main() {
	var analizator []*analysis.Analyzer
	analizator = append(analizator,
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		framepointer.Analyzer,
		hostport.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		slog.Analyzer,
		stdmethods.Analyzer,
		stdversion.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		waitgroup.Analyzer,
	)
	for _, v := range staticcheck.Analyzers {
		analizator = append(analizator, v.Analyzer)
	}
	for _, v := range simple.Analyzers {
		if v.Analyzer.Name == "S1002" {
			analizator = append(analizator, v.Analyzer)
		}
	}
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1003" {
			analizator = append(analizator, v.Analyzer)
		}
	}
	for _, v := range quickfix.Analyzers {
		if v.Analyzer.Name == "QF1003" {
			analizator = append(analizator, v.Analyzer)
		}
	}
	analizator = append(analizator, errcheck.Analyzer)
	analizator = append(analizator, bodyclose.Analyzer)
	analizator = append(analizator, osExitCheckAnalyzer)
	multichecker.Main(analizator...)
}
