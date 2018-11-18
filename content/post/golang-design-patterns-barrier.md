---
title: "Golang Wzorce Projektowe - Bariera"
date: 2018-09-13T20:49:25+02:00
draft: false
image: "img/barrier.png"
description = "TODO"
---

## Wzorzec

Bariera jest to jeden ze wzorców programowania współbieżnego. 
Pozwala on zsynchronizować wykonywanie się kilku wątków/procesów/współprogramów (ang. coroutine; dla uproszczenia nazywajmy je procesami) w taki sposób, że wszystkie one muszą dotrzeć do wskazanego miejsca (bariery), zanim będą mogły wykonywać się dalej.

Jako przykład zastosowania można podać kilku etapowy algorytm operujący równolegle na współdzielonej macierzy.
W takim przypadku bariera pomogłaby zagwarantować, że wszystkie procesy
zaktualizowały współdzielony stan w fazie `t`.
Tak, żeby w fazie `t+1` te same dane mogłyby być użyte jako dane wejściowe.

Bariera może być zaimplementowana na wiele sposobów.
Różnić się ona może zastosowanymi strukturami danych: licznik, drzewo.
Sposobem realizacji: [wirująca](https://en.wikipedia.org/wiki/Busy_waiting) (ang. spinning),
przy pomocy [dyspozytora](https://pl.wikipedia.org/wiki/Dyspozytor) (ang. scheduler).
Każda implementacja adresuje inny problem oraz inną architekturę procesora.

## Implementacja

Chciałbym porównać kilka sposobów implementacji bariery przy pomocy dostępnych w **Golang**'u mechanizmów synchronizacji.
Wszystkie one będą spełniać takie same założenia:

* Liczba procesów jest stała i znana odgórnie.  
* Raz utworzona bariera powinna móc być użyta wielokrotnie.

Ponadto każda implementacja musi spełniać jeden poniższych interfejsów:

```go
type GenericBarrier interface {
	Await()
}

type CSPBarrier interface {
	Await() <-chan struct{}
}
```
 
Przykładowa aplikacja bariery mogłaby wyglądać następująco:

```go
bar := barrier.New(concurrency)
for i :=0;i<steps;i++{
    for j := 0; j<concurrency; j++ {
        go func(){
            // Logic before.
            bar.Await()
            // Logic after.
        }()
    }
    // Wait for the result.
}
```

![binary heap](/img/barrier.png#center)

### CSP

Model [Communicating Sequential Processes](https://en.wikipedia.org/wiki/Communicating_sequential_processes) był wzorem dla twórcow Go podczas projektowania współbierznej natury języka.
W efekcie otrzymaliśmy wysokopoziomową abstrakcję w postaci kanałów (ang. channels).
Można go streścić tymi słowami:

> Do not communicate by sharing memory; instead, share memory by communicating.

Jak łatwo się dymyśleć, implementacja w oparciu o kanały nie będzie wyróżniać się na tle reszty oszałamiającą wydajnością.
Będzie jednak miała inną bardzo cenną własność.
Ze względu na urzycie kanałów, implementacja ta będzie się dała niezwykle łatwo komponować z resztą sygnałów przy pomocy `select`.

```go
type Barrier struct {
	concurrency int
	iteration   chan uint8
	waiting     [2]chan struct{}
	in          chan struct{}
}

func NewBarrier(i int) *Barrier {
	iteration := make(chan uint8, i)
	for j := 0; j < i; j++ {
		iteration <- 0
	}
	return &Barrier{
		concurrency: i,
		in:          make(chan struct{}, i-1),
		waiting: [2]chan struct{}{
			make(chan struct{}),
			nil,
		},
		iteration: iteration,
	}
}
```

* `concurrency` - Odpowiada za przechowywanie liczby procesów, które ta bariera będzie chronić.
* `iteration` - Przyjmuje wartości `0` oraz `1`. Przedstawia parzystę i nieparzyste przebiegi.
* `waiting` - Tablica dwóch kanałów zwracanych na przemiennie (co iterację) przez metodę `Await`. 
* `in` - Kanał pełniący rolę bufora. Jego zapełnienie lub całkowite zwolnienie oznacza osiągniecie bariery przez wszystkie procesy. 

```go
func (b *Barrier) Await() <-chan struct{} {
	i := <-b.iteration
	b.iteration <- (i + 1) % 2

	if i != 1 {
		select {
		// Write into the channel so, ...
		case b.in <- struct{}{}:
			return b.waiting[i%2]
		default:
			return b.reset(i)
		}
	}

	select {
	// ... during even iteration, it can be drained.
	case <-b.in:
		return b.waiting[i%2]
	default:
		return b.reset(i)
	}
}

func (b *Barrier) reset(i uint8) <-chan struct{} {
	b.waiting[(i+1)%2] = make(chan struct{})
	close(b.waiting[i%2])

	return b.waiting[i%2]
}
```

Wyróżnić możemy dwa różne przebiegi, parzyste oraz nieparzyste.
Wczasie tego drugiego kanał `in` oraz `iteration` jest wypełniany.
Gdy tylko `in` osiąnie swój limit, ostatnio zwrócony kanał `waiting` zostanie zamknięty. 
Co da sygnał nasłuchującym procesom że blokada została zwolniona.
W momencie gdy kanał `iteration` zacznie informować o przebiegu parzystym kanał `in` zacznie być stopniowo opróżniany.

Dzięki zastosowaniu kanałów przenoszących `struct{}`, metoda `Await` nie powoduje dodatkowych alokacji.

### Monitor
Paczka [sync](https://golang.org/pkg/sync/) dostarcza nam szereg wartościowych narzędzi, które adresują problem synchronizacji.
Jednym z nich jest struktura [sync.Cond](https://golang.org/pkg/sync/#Cond).
Jest ona implementacją wzorca [Monitor](https://pl.wikipedia.org/wiki/Monitor_(programowanie_współbieżne)), który pozwala powiadomić zainteresowane procesy momencie, gdy dany warunek zostanie spełniony.
W tym przypadku będzie to liczba procesów, które osiągnęły wskazane miejsce w programie.

```go
type Barrier struct {
	c                      *sync.Cond
	concurrency, iteration int
	waiting                [2]int
	broadcasted            [2]bool
}

func NewBarrier(n int) *Barrier {
	return &Barrier{
		c:           sync.NewCond(&sync.RWMutex{}),
		concurrency: n,
	}
}
```

Podobnie jak w przypadku implementacji w oparciu o kanały, przebiegi parzyste i nieparzyste będą śledzone osobno.

* `concurrency` - Odpowiada za przechowywanie liczby procesów, które ta bariera będzie chronić.
* `iteration` - Liczba razy, kiedy wszystkie procesy osiągnęły barierę.
* `waiting` - Ilości procesów, które osiągnęły wyznaczony punkt w czasie jednej (obecnej i poprzedniej) iteracji. 
* `broadcasted` - Flaga informująca czy dla danej iteracji sygnał został już rozesłany do nasłuchujących procesów. 


```go
func (b *Barrier) Await() {
	b.c.L.Lock()
	i := b.iteration
	b.waiting[i%2]++
	for b.concurrency != b.waiting[i%2] {
		b.c.Wait()
	}
	if !b.broadcasted[i%2] {
		b.broadcasted[i%2] = true
		b.broadcasted[(i+1)%2] = false
		b.waiting[(i+1)%2] = 0
		b.iteration++
		b.c.Broadcast()
	}

	b.c.L.Unlock()
}
```  
Metoda `Await` składa się z dwóch części.
Pierwszej, zwiększającej licznik i wprowadzającej proces w stan uśpienia.
Oraz drugiej, resetującej stan zmiennych dla kolejnego przebiegu oraz rozsyłającej sygnały o opuszczeniu bariery. 

Warto tutaj zwrócić uwagę na pętlę `for`. W przypadku zmiennych warunkowych istnieje szansa na spontaniczne wybudzenie (ang. [spurious wakeup](https://en.wikipedia.org/wiki/Spurious_wakeup)).
Z tego powodu `Wait` musi być umiejscowiony w pętli, tak żeby warunek był sprawdzany po każdym wybudzeniu, także tym fałszywym.

### CAS

Jednym ze sposobów synchronizacji dostępu do danych jest zastosowanie wsparcia sprzętowego w postaci operacji [Compare-and-swap](https://pl.wikipedia.org/wiki/Compare-and-swap).
Operacja ta jest zaimplementowana w **Golang**'u w paczce [sync/atomic](https://golang.org/pkg/sync/atomic/).
Paczka implementuje kilka wariantów dla różnych typów prostych.

```go
type Barrier struct {
	concurrency uint64
	iteration   uint64
	waiting     [2]uint64
}

func NewBarrier(i int) *Barrier {
	return &Barrier{
		concurrency: uint64(i),
		iteration:   1,
		waiting: [2]uint64{
			uint64(i),
			uint64(i),
		},
	}
}
```

* `concurrency` - Odpowiada za przechowywanie liczby procesów, które ta bariera będzie chronić.
* `iteration` - Liczba razy, kiedy wszystkie procesy osiągnęły barierę.
* `waiting` - Dwuelementowa tablica przetrzymująca informację o ilości procesów, które osiągnęły wyznaczony punkt w czasie jednej (obecnej i poprzedniej) iteracji. 

Za każdym razem, gdy proces osiągnie wyznaczony punkt i zostanie wywołana metoda `Await` licznik dla danej iteracji zostanie zdekrementowany.
Zaraz potem proces przejdzie w stan oczekiwania.
Proces będzie czekał (w nieskończonej pętli), do momentu aż wartość licznika zejdzie do zera.

Istotnym elementem tej implementacji jest dyspozytor współprogramów (ang. coroutine scheduler).
Za pomocą funkcji `runtime.Gosched()` jesteśmy w stanie wydać mu polecenie, aby ten przełączył się na inny proces.
Bez tego mechanizmu implementacja byłaby niemożliwa.
Liczba wywołań `Await` nie mogłaby przekroczyć liczby procesorów.
Czyniłoby to tę implementację zwyczajnie bezużyteczna.
 
```go
func (b *Barrier) Await() {
	i := atomic.LoadUint64(&b.iteration)

	if atomic.AddUint64(&b.counter[i%2], ^uint64(0)) == 0 {
		return
	}

	for {
		runtime.Gosched()

		switch atomic.LoadUint64(&b.counter[i%2]) {
		case b.concurrency: // 1
			return
		case 0: // 2
			atomic.AddUint64(&b.iteration, 1)
			atomic.StoreUint64(&b.counter[i%2], b.concurrency)

			return
		}
	}
}
```

Z uwagi na to, że nie mamy kontroli nad tym, kiedy proces zostanie znowu wznowiony przez dyspozytora.
Jest wielce prawdopodobne, że proces, który ostatni osiągnie punkt kontrolny, nie będzie ostatnim, który go opuści.
Jest także całkowicie normalną sytuacją, gdzie część procesów czeka, aby opuścić poprzednią barierę, gdy druga część już osiągnęła następną.
Z tego właśnie powodu wartość licznika jest porównywana do `b.concurrency` a sam licznik składa się z dwóch elementów.
Umożliwia to procesom, które jeszcze znajdują się w poprzedniej iteracji skutecznie ją opuścić.

### Wydajność

W zależności od sposobu wykorzystania bariery, będzie nas interesować inny sposób mierzenia wydajności.
W przypadku krótkich cyklów, gdzie bariera jest często alokowana i jest wykorzystywana przez krótki czas. 
Będzie nas interesować koszt utworzenia (jak i usunięcia poprzez GC).

Gdy jednak nasz przypadek przewiduje tworzenie bariery z żadka, będziemy mogli się skupić jedynie na wydajności metody `Await`. 

#### Konstruktor i Await

TODO 

##### CSP

```bash
BenchmarkBarrier/4-4       	  500000	      2238 ns/op	     352 B/op	       4 allocs/op
BenchmarkBarrier/8-4       	  500000	      3777 ns/op	     352 B/op	       4 allocs/op
BenchmarkBarrier/16-4      	  200000	      6029 ns/op	     352 B/op	       4 allocs/op
BenchmarkBarrier/32-4      	  200000	     10615 ns/op	     368 B/op	       4 allocs/op
BenchmarkBarrier/64-4      	  100000	     19209 ns/op	     400 B/op	       4 allocs/op
BenchmarkBarrier/128-4     	   50000	     32745 ns/op	     464 B/op	       4 allocs/op
BenchmarkBarrier/256-4     	   20000	     63591 ns/op	     592 B/op	       4 allocs/op
BenchmarkBarrier/512-4     	   10000	    123848 ns/op	     884 B/op	       4 allocs/op
BenchmarkBarrier/1024-4    	    5000	    247242 ns/op	    1410 B/op	       4 allocs/op
BenchmarkBarrier/2048-4    	    3000	    533431 ns/op	    2665 B/op	       5 allocs/op
BenchmarkBarrier/4096-4    	    2000	   1146631 ns/op	    5504 B/op	       8 allocs/op
```

##### Monitor

```bash
BenchmarkBarrier/4-4       	 1000000	      1859 ns/op	     144 B/op	       3 allocs/op
BenchmarkBarrier/8-4       	  500000	      2824 ns/op	     144 B/op	       3 allocs/op
BenchmarkBarrier/16-4      	  300000	      4440 ns/op	     144 B/op	       3 allocs/op
BenchmarkBarrier/32-4      	  200000	      7997 ns/op	     144 B/op	       3 allocs/op
BenchmarkBarrier/64-4      	  100000	     14201 ns/op	     144 B/op	       3 allocs/op
BenchmarkBarrier/128-4     	   50000	     26508 ns/op	     144 B/op	       3 allocs/op
BenchmarkBarrier/256-4     	   30000	     49453 ns/op	     144 B/op	       3 allocs/op
BenchmarkBarrier/512-4     	   20000	     99497 ns/op	     145 B/op	       3 allocs/op
BenchmarkBarrier/1024-4    	   10000	    200991 ns/op	     149 B/op	       3 allocs/op
BenchmarkBarrier/2048-4    	    3000	    448071 ns/op	     178 B/op	       3 allocs/op
BenchmarkBarrier/4096-4    	    2000	    945275 ns/op	     270 B/op	       4 allocs/op
```

##### CAS

```bash
BenchmarkBarrier/4-4       	 1000000	      1561 ns/op	      32 B/op	       1 allocs/op
BenchmarkBarrier/8-4       	 1000000	      2159 ns/op	      32 B/op	       1 allocs/op
BenchmarkBarrier/16-4      	  300000	      4201 ns/op	      32 B/op	       1 allocs/op
BenchmarkBarrier/32-4      	  200000	      7765 ns/op	      32 B/op	       1 allocs/op
BenchmarkBarrier/64-4      	  100000	     14110 ns/op	      32 B/op	       1 allocs/op
BenchmarkBarrier/128-4     	   50000	     25768 ns/op	      32 B/op	       1 allocs/op
BenchmarkBarrier/256-4     	   30000	     52309 ns/op	      32 B/op	       1 allocs/op
BenchmarkBarrier/512-4     	   10000	    105654 ns/op	      32 B/op	       1 allocs/op
BenchmarkBarrier/1024-4    	    5000	    217423 ns/op	      32 B/op	       1 allocs/op
BenchmarkBarrier/2048-4    	    3000	    438460 ns/op	      33 B/op	       1 allocs/op
BenchmarkBarrier/4096-4    	    2000	   1001289 ns/op	      33 B/op	       1 allocs/op
```

TODO 

#### Await

TODO

##### CSP

```bash
BenchmarkBarrier_Await/4-4 	         1000000	      1154 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/8-4 	         1000000	      1800 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/16-4         	  500000	      3406 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/32-4         	  200000	      6511 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/64-4         	  100000	     12926 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/128-4        	   50000	     25473 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/256-4        	   30000	     47043 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/512-4        	   20000	     89603 ns/op	       4 B/op	       0 allocs/op
BenchmarkBarrier_Await/1024-4       	   10000	    185166 ns/op	      16 B/op	       0 allocs/op
BenchmarkBarrier_Await/2048-4       	    3000	    423602 ns/op	     113 B/op	       1 allocs/op
BenchmarkBarrier_Await/4096-4       	    2000	    894780 ns/op	     518 B/op	       5 allocs/op
```

##### Monitor

```bash
BenchmarkBarrier_Await/4-4 	         1000000	      1037 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/8-4 	         1000000	      1729 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/16-4         	  500000	      3482 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/32-4         	  200000	      6935 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/64-4         	  100000	     13125 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/128-4        	   50000	     25690 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/256-4        	   30000	     48422 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/512-4        	   20000	     94228 ns/op	       1 B/op	       0 allocs/op
BenchmarkBarrier_Await/1024-4       	   10000	    194547 ns/op	       3 B/op	       0 allocs/op
BenchmarkBarrier_Await/2048-4       	    5000	    426183 ns/op	      23 B/op	       0 allocs/op
BenchmarkBarrier_Await/4096-4       	    2000	    987861 ns/op	     134 B/op	       1 allocs/op
```

##### CAS

```bash
BenchmarkBarrier_Await/4-4 	         1000000	      1104 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/8-4 	         1000000	      1785 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/16-4         	  500000	      3713 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/32-4         	  200000	      6884 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/64-4         	  100000	     12872 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/128-4        	   50000	     24538 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/256-4        	   30000	     50684 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/512-4        	   10000	    101821 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/1024-4       	   10000	    211846 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarrier_Await/2048-4       	    3000	    466017 ns/op	       1 B/op	       0 allocs/op
BenchmarkBarrier_Await/4096-4       	    2000	   1029696 ns/op	       0 B/op	       0 allocs/op
```

TODO
