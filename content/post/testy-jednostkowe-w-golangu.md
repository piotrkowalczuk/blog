+++
category = ["tutorial"]
title = "Testy jednostkowe w Golangu"
description = "Testowanie przykładowej aplikacji internetowej w sposób czytelny i ustandaryzowany."
date = 2018-02-18T20:11:00+01:00
draft = false
tags = ["golang", "mock", "stub", "test", "unit"]
keywords = [
	"golang", "mock", "stub", "test", "unit", "mockery"
]
+++

## Wstęp
Testowanie jednostkowe to jedna z podstawowych technik weryfikowania poprawnego działania programu. Nie oznacza to jednak, że temat jest prosty. Szczególnie w przypadku  __Go__, gdzie biblioteka [testing](https://golang.org/pkg/testing/), mimo iż potężna, nie narzuca jednego właściwego podejścia do tematu. Daje to nam dużą swobodę, ale nie za darmo. W przypadku większych zespołów ta swoboda może być problemem. Warto się wtedy zastanowić nad ustandaryzowaniem swojego podejścia.

Chciałbym się podzielić z wami moim sposobem pisania nieco bardziej złożonych testów jednostkowych. Nie będzie to nic wyrafinowanego. Celem nadrzędnym jest, aby po spotkaniu z nieznanym do tej pory kodem, interpretowanie oraz rozszerzanie testów było proste.

## Problem
Przyjmijmy, że mamy do przetestowania kontroler naszego web serwisu. 
Jest to prosta aplikacja, która umożliwia dilerowi samochodów zarządzanie swoją flotą.

### Kontroler
Kontrolery w naszej aplikacji muszą spełniać następujący interfejs:

```go
type controller interface {
	handle(*http.Request) (interface{}, error)
}
```

Taka abstrakcja pozwala nam, odseparować warstwę biznesową od transportowej.
Ktoś mógłby zwrócić uwagę, że przez użycie `http.Request` jest to niemożliwe. Aby nie komplikować naszego przykładu aż zanadto, musimy zaakceptować to niewielkie niedociągnięcie.

Na tapetę weźmiemy kontroler dodawania oraz modyfikowania samochodów, którego uproszczona implementacja mogłaby wyglądać następująco:

```go
package example 

type Payload struct {
	ID          int64 `storage:"identifier"` // if set, storage will perform update, otherwise insert
	Name        string
	Age, Mileage int
	Owner       string
}

type PutCarController struct {
	Storage Storage
}

func (pcc *PutCarController) Handle(req *http.Request) (interface{}, error) {
	var pay Payload
	if err := json.NewDecoder(req.Body).Decode(&pay); err != nil {
		return nil, err
	}
	req.Body.Close()

	if pay.Name == "" {
		return nil, errors.New("missing name")
	}

	if err := pcc.Storage.Put(req.Context(), &pay); err != nil {
		return nil, err
	}

	return &pay, nil
}
```

Jest ona pozbawiona wszelkiego rodzaju ozdobników. Nawet walidacja żądania jest uproszczona. To, co sprawi najwięcej problemu podczas testowania tego kawałka kodu to przygotowanie [atrapy](https://pl.wikipedia.org/wiki/Atrapa_obiektu) bazy danych.
 
### Baza danych
Jaka jest to baza danych, nie ma dla nas żadnego znaczenia. Chociaż nie ukrywam, że planując jej interfejs, wzorowałem się na [Google Datastore](https://cloud.google.com/datastore/docs/concepts/overview). Oto on:

```go
package example 

type Storage interface {
	Put(context.Context, interface{}) error
	Get(context.Context, int64) (interface{}, error)
}
```

`Put` zapisuje obiekt do bazy danych. Jeżeli operacja zakończy się sukcesem i obiekt nie miał wcześniej nadanego identyfikatora (pole oznaczona tagiem `identifier`), zostanie mu on nadany, a przekazany obiekt zmodyfikowany o ten identyfikator. W razie fiaska zwraca błąd. 

`Get` nie jest nam do niczego potrzebny, jest tutaj jedynie, aby nadać sensu kolejnej sekcji ;)

## Stub czy mock?
Nasz przypadek jest bardzo uproszczony. Użycie stuba wydaje się (i słusznie) uzasadnione. Oto jak taki stub mógłby wyglądać:

```go
type testStorage struct {
	storage // embeded interface
	
	id int64
	err error
}

func (ts *testStorage) Put(_ context.Context, obj interface{}) (int64, error) {
	return ts.id, ts.err
}
```

Dzięki osadzeniu interfejsu `storage` w `testStorage` nasza struktura implementuje cały potrzebny interfejs. 
Trzeba jednak pamiętać, że wywołanie `Get` zakończy się, wyrzuceniem wyjątku (panic).

Aby spełnić obietnicę z tytułu, przekombinujemy nieco nasze rozwiązanie. 
Nie zważając na rozsądek, wykorzystamy mocki.

Do utworzenia atrap, posłuży nam [mockery](https://github.com/vektra/mockery).
Narzędzie te w połączeniu z `go generate` umożliwi nam w łatwy sposób wygenerować wszystkie potrzebne obiekty. 
W bardziej złożonych aplikacjach takie podejście odpłaci się z nawiązką.

Możemy ten proces zautomatyzować, dodając do naszego kodu:

```go
//go:generate mockery -case=underscore -all
```

Dzięki temu, przy każdorazowym wywołaniu komendy `go generate`, wszystkie atrapy zostaną wygenerowane automatycznie.

## Scenariusz
Nasz test powinien pokrywać możliwie dużo pozytywnych, jak i negatywnych przypadków. Powinny być one, od siebie całkowicie odseparowane (nie mogą dzielić stanu). Dodawanie nowych przypadków powinno być proste i nie narażać już istniejących na modyfikacje. 

### Table Driven Testing
Jest to powszechnie stosowany wzorzec, polegający na grupowaniu różnych przypadków w jednym teście i iterowaniu po nich. Przeciwieństwem jest tworzenie osobnego testu dla każdego przypadku z osobna:

```go
func TestPutCarController_Handle_success(t *testing.T) { ... }
func TestPutCarController_Handle_deadlineExceeded(t *testing.T) { ... }
```

[TDT](https://github.com/golang/go/wiki/TableDrivenTests) ułatwia utrzymywanie oraz poruszanie się po testach. Osobiście jestem zwolennikiem stosowania mapy, gdzie klucze służą wyjaśnieniu co taki test ma udowodnić oraz pozwalają szybko przeskoczyć (cmd+f) z lini komend do konkretnego miejsca w kodzie, gdzie taki test jest zdefiniowany. 

Aby przetestować metodę `Handle` naszego kontrolera, będziemy potrzebować struktury opisującej kolekcje przypadków.

```go
cases := map[string]struct {
	req  *http.Request
	init func(*testing.T)
	res  interface{}
	err  error
}{}
```

Składa się ona z:

* `req` - argumentu przekazywanego do metody `Handle`
* `init` - funkcji inicjalizującej wszystkie atrapy, może być `nil`
* `res` - przewidywanej odpowiedzi
* `err` - w razie, jeżeli jest to scenariusz testujący pesymistyczny przypadek, potrzebujemy obiektu błędu do porównania

### Szablon
Szablon, na razie bez zaimplementowanych przypadków, wygląda następująco:

```go
package example_test

func TestPutCarController_Handle(t *testing.T) {
	storage := &mocks.Storage{}

	req := func(t *testing.T, obj interface{}) *http.Request {
		buf, err := json.Marshal(obj)
		if err != nil {
			t.Fatalf("payload marshal failure: %s", err.Error())
		}
		return httptest.NewRequest(
			http.MethodPut, 
			"http://localhost", 
			bytes.NewReader(buf),
		)
	} // 1
	
	cases := map[string]struct {
		req  *http.Request
		init func(*testing.T)
		res  interface{}
		err  error
	}{
		// TODO: implement
	}

	for hint, given := range cases {
		t.Run(hint, func(t *testing.T) {
			storage.ExpectedCalls = nil // 2

			if given.init != nil { // 3
				given.init(t)
			}

			got, err := (&example.PutCarController{
				Storage: storage,
			}).Handle(given.req)
			if given.err != nil { // 4
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if given.err.Error() != err.Error() {
					t.Fatalf("wrong error, expected:\n	%s\nbut got:\n	%s", given.err.Error(), err.Error())
				}
			} else {
				if !reflect.DeepEqual(given.res, got) {
					t.Fatalf("wrong response, expected:\n	%v\nbut got:\n	%v", given.res, got)
				}
			}

			mock.AssertExpectationsForObjects(t, storage) // 5
		})
	}
}
```

Warto się na chwile pochylić nad powyższym kodem i przeanalizować go nieco głębiej. Snippet ten posiada kilka oznaczonych punktów, które są warte wyjaśnienia:

1. Pomocnicza funkcja, która inicjalizuje obiekt `http.Request`, który niesie ze sobą dane w formacie JSON.
2. Resetowanie mock'ów.
3. Nie każdy przypadek będzie potrzebował dodatkowej inicjalizacji. Dla utrzymania przejrzystości, funkcja `init` jest opcjonalna.
4. Sprawdzamy, czy oczekiwanym rezultatem jest błąd. Innymi słowy, czy jest to przypadek pesymistyczny. Jeżeli tak, porównujemy zwrócony błąd, do tego którego oczekujemy. 
5. Sprawdzamy, czy liczba wywołań metod naszych atrap zgadza się z oczekiwaniami.

### Przypadki

#### Brakująca nazwa

```go
cases := map[string]struct {
	req  *http.Request
	init func(*testing.T)
	res  interface{}
	err  error
}{
	"missing-name": {
		req: req(t, &example.Payload{}),
		err: errors.New("missing name"),
	},
}
```

#### Przekroczenie czasu żądania
Ten test weryfikuje, czy zwrócony błąd przez bazę danych jest przekazany dalej. Wspólny dekorator dla wszystkich kontrolerów mógłby interpretować taki błąd i ustawiać odpowiedni kod statusu. W tym przypadku byłby to `http.StatusRequestTimeout`.

```go
cases := map[string]struct {
	req  *http.Request
	init func(*testing.T)
	res  interface{}
	err  error
}{
	"deadline-exceeded": {
		req: req(t, &example.Payload{
			Name: "brand new car",
		}),
		err: context.DeadlineExceeded,
		init: func(t *testing.T) {
			storage.On("Put", mock.Anything, mock.AnythingOfType("*example.Payload")).
				Return(context.DeadlineExceeded).
				Once()
		},
	},
}
```

#### Zły format żądania
W przypadku gdy klient wyśle źle sformatowane dane, powinien zostać powiadomiony o tym.

```go
cases := map[string]struct {
	req  *http.Request
	init func(*testing.T)
	res  interface{}
	err  error
}{
	"text-payload": {
		req: httptest.NewRequest(http.MethodPut, "http://localhost", strings.NewReader("not-json-at-all")),
		err: errors.New("invalid json payload"),
	},
}
```

#### Sukces
Zdecydowanie najpełniejszy przykład. Pokazuje on jak stosować paczkę [mock](https://godoc.org/github.com/stretchr/testify/mock).

```go
cases := map[string]struct {
	req  *http.Request
	init func(*testing.T)
	res  interface{}
	err  error
}{
	"success": {
		req: req(t, &example.Payload{
			Name: "brand new car",
		}),
		res: &example.Payload{
			ID:   100,
			Name: "brand new car",
		},
		init: func(t *testing.T) {
			storage.On("Put", mock.Anything, mock.AnythingOfType("*example.Payload")).
				Run(func(args mock.Arguments) {
					if pay, ok := args.Get(1).(*example.Payload); ok {
						pay.ID = 100
					}
				}).
				Return(nil).
				Once()
		},
	},
}
```

## Podsumowanie 
Prezentowany sposób jest czytelny i nieźle skaluje się wraz ze wzrostem przypadków, jak i testów. Osiągnięcie pełnego pokrycia kodu w testach nie jest problemem: 

```
go test -v -cover
=== RUN   TestPutCarController_Handle
=== RUN   TestPutCarController_Handle/deadline-exceeded
=== RUN   TestPutCarController_Handle/text-payload
=== RUN   TestPutCarController_Handle/missing-name
=== RUN   TestPutCarController_Handle/success
--- PASS: TestPutCarController_Handle (0.00s)
    --- PASS: TestPutCarController_Handle/deadline-exceeded (0.00s)
    	mock.go:420: PASS:	Put(string,mock.AnythingOfTypeArgument)
    --- PASS: TestPutCarController_Handle/text-payload (0.00s)
    --- PASS: TestPutCarController_Handle/missing-name (0.00s)
    --- PASS: TestPutCarController_Handle/success (0.00s)
    	mock.go:420: PASS:	Put(string,mock.AnythingOfTypeArgument)
PASS
coverage: 100.0% of statements
ok  	github.com/piotrkowalczuk/blog/examples/testy-jednostkowe-w-golangu	0.016s
```

Jest jednak jeszcze trochę miejsca na ulepszenia. Wspólna funkcja `assertError` pozwoliłoby wyeliminować duplikację kodu pomiędzy testami. Dodanie własnego typu  błędu umożliwiłoby lepszą obsługę błędów w samym kontrolerze, jak i bardziej elastyczną asercję. Dekodowanie zawartości żądania poza kontrolerem pozwoliłoby na jeszcze lepszą separację warstw. 

W kolejnym kroku polecam zapoznać się ze świetnym wpisem [Error handling in Upspin](https://commandcenter.blogspot.de/2017/12/error-handling-in-upspin.html). Rob Pike przedstawia tam dość nowatorskie jak na standardy Go podejście do obsługi błędów.

Pełen kod źródłowy do wglądu [tutaj](https://github.com/piotrkowalczuk/blog/tree/master/examples/testy-jednostkowe-w-golangu).




