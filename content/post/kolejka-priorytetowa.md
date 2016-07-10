+++
category = ["przykłady"]
date = "2016-07-10T20:12:00+02:00"
tags = ["struktury danych", "golang", "biblioteka standardowa"]
title = "Kolejka Priorytetowa"
description = "Implementacja kolejki priorytetowej przy pomocy bibliteki standardowej na przykładzie kolejki poleceń"
+++

# Wstęp

Kolejka priorytetowa to abstrakcyjna struktura danych gdzie elementy są uszeregowane według danej wielkości.
Kolejka ta nie jest kolejką typu FIFO czy też LIFO.
Ponieważ jest to [ADT](https://pl.wikipedia.org/wiki/Abstrakcyjny_typ_danych) może być ona zaimplementowana na wielę sposobów:

* Kopiec binarny
* Kopiec dwumianowy
* Tablica
* Lista
* Rownoważone BST
* Kopiec Fibonacciego
* Kolejka Brodala

# Zagadnienie

Aby nadać problemowi bardziej intuicyjny charakter spróbujmy zastosować go w praktyce.
Naszym zadaniem będzie utworzenie kolejki poleceń. Zadania mogą być dodawane w dowolnej kolejności.
Każde polecenie powinno być opisane przez:

* `id` - identyfikator
* `index` - miejsce w kolejce
* `name` - nazwa
* `timestamp` - data do kiedy zadanie powinno zostać wykonane
* `epsilon` - czas po wyznaczonej dacie wykonania po upłynięciu którego zadanie powinno zostać porzucone
* `command` - komenda uruchamiająca skrypt/program

## Propozycja rozwiązania

Jak wszyscy wiemy, biblioteka standardowa Go jest naprawdę bogata.
Nie zabrakło także implementacji stosu (binarnego).
Paczka [heap](https://golang.org/pkg/container/heap/), bo o niej tutaj mowa dostarcza nam taki oto interface:

```go
type Interface interface {
	sort.Interface
	Push(x interface{})
	Pop() interface{}
}
```

Który jeżeli zaimplementowany poprawnie może być wykorzystany przy użyciu szeregu funkcji:

* [heap.Init](https://golang.org/pkg/container/heap/#Init) - inicjalizacja stosu, `O(n)`
* [heap.Push](https://golang.org/pkg/container/heap/#Push) - dodanie elementu, `O(log(n))`
* [heap.Pop](https://golang.org/pkg/container/heap/#Pop) - usuniecie elementu minimalnego, `O(log(n))`
* [heap.Remove](https://golang.org/pkg/container/heap/#Remove) - usuniecie elementy pod podanym indexem, `O(log(n))`
* [heap.Fix](https://golang.org/pkg/container/heap/#Fix) - "naprawienie" stosu po np zmianie wartosci jednego z elementow, `O(log(n))`

## Implementacja

Raz jeszcze skorzystamy z dobrodziejstw biblioteki standardowej.
Znaleźć w niej możemy paczkę [exec](https://golang.org/pkg/os/exec/) w której to znajduje się struktura [exec.Cmd](https://golang.org/pkg/os/exec/#Cmd).
Struktura reporezentująca pojedyńczą pracę do wykonania mogłaby wyglądać następująco:

```go
type Job struct {
	ID int64
	Index int64
	Name string
	Timestamp time.Time
	Epsilon time.Duration
	Command *exec.Cmd
}

type Jobs []*Job
```

Colekcja takich struktur musi implementować wcześniej wymieniony [Interface](https://golang.org/pkg/container/heap/#Interface).
A więc po kolei:

### Len

Trywialna metoda, która jedynie sprawdza długość kolekcji.

```go
// Len implements sort Interface.
func (j Jobs) Len() int {
	return len(j)
}
```

### Less

Metoda ta nie tylko sprawdza który deadline nastąpi jako pierwszy,
ale w wypadku gdy są równe, porównuje także wartość `Epsilon`.
Implementacja ta to jedynie przykład.


```go
// Less implements sort Interface.
func (j Jobs) Less(n, m int) bool {
	if j[n].Timestamp.Equal(j[m].Timestamp) {
		return j[n].Epsilon < j[m].Epsilon
	}
	return j[n].Timestamp.Before(j[m].Timestamp)
}
```

### Swap

Nic innego jak zamiana elementów pod wskazanymi indeksami.

```go
// Swap implements sort Interface.
func (j Jobs) Swap(n, m int) {
	j[n], j[m] = j[m], j[n]
	j[n].Index = int64(n)
	j[m].Index = int64(m)
}
```


### Push
Przez push rozumiemy dodanie na koniec kolekcji nowego elementu i ustawienie jego indeksu na `n`.
Funkcja [heap.Push](https://golang.org/pkg/container/heap/#Push) użyje tej metody na początku, a następnie będzie przesuwać element do góry tak długo jak to tylko możliwe, aż osiągnie właściwy dla siebie index wynikający z warunku zawartego w metodzie `Less`.
```go
// Push implements heap Interface.
func (j *Jobs) Push(x interface{}) {
	n := len(*j)
	item := x.(*Job)
	item.Index = int64(n)
	*j = append(*j, item)
}
```

### Pop

Usuniecie zadania z kolejki odbywa się przez utworzenie nowego [slice'a](https://blog.golang.org/go-slices-usage-and-internals) z pominięciem ostatniego elementu.
Dla jasności ustawiamy także `Index` usuniętego elementu na `-1`.

```go
// Pop implements heap Interface.
func (j *Jobs) Pop() interface{} {
	old := *j
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*j = old[0 : n-1]
	return item
}
```

### Efekt końcowy

Całość działa tak jak przewiduje to koncepcja kolejki priorytetowej. [Testowy przykład](https://blog.golang.org/examples) to potwierdza.
```go
func ExampleJobs() {
	jobs := make(Jobs, 0, 3)
	zero := time.Now()
	heap.Init(&jobs)
	heap.Push(&jobs, &Job{
		ID:        1,
		Name:      "first",
		Timestamp: zero.Add(10 * time.Hour),
		Epsilon:   5 * time.Minute,
		Command:   exec.Command("ls", "-lha"),
	})
	heap.Push(&jobs, &Job{
		ID:        1,
		Name:      "second",
		Timestamp: zero.Add(10 * time.Hour),
		Epsilon:   4 * time.Minute,
		Command:   exec.Command("pwd"),
	})
	heap.Push(&jobs, &Job{
		ID:        1,
		Name:      "third",
		Timestamp: zero.Add(5 * time.Hour),
		Epsilon:   4 * time.Minute,
		Command:   exec.Command("ps", "aux"),
	})
	fmt.Println(jobs.Len())

	j1 := heap.Pop(&jobs).(*Job)
	j2 := heap.Pop(&jobs).(*Job)
	j3 := heap.Pop(&jobs).(*Job)

	fmt.Println(j1.Name)
	fmt.Println(j2.Name)
	fmt.Println(j3.Name)
	fmt.Println(jobs.Len())

	// Output:
	// 3
	// third
	// second
	// first
	// 0
}
```

# Wnioski

Kolejka priorytetowa to bardzo ważna struktura danych szczególnie w przypadku wszelkiego rodzaju aplikacji serwerowych.
Przed przystąpieniem do rozwiazania problemu warto przyjrzeć się [bibliotece standardowej](https://golang.org/pkg/),
może zawierać ona przydatne interfacy/implementacje czy także [przykłady](https://golang.org/pkg/container/heap/#example__priorityQueue):

* [container/heap](https://golang.org/pkg/container/heap) - implementacja stosu
* [container/list](https://golang.org/pkg/container/list) - implementacja podwujnie łączonej listy
* [container/ring](https://golang.org/pkg/container/ring) - operacje na listach cyklicznych