+++
category = ["opinie"]
date = "2017-10-21T19:56:18+02:00"
lastmod = "2017-11-19T00:12:55+01:00"
title = "W poszukiwaniu pracy jako programista Golang"
tags = ["golang", "praca", "rekrutacja"]
description = "Rekrutacja na stanowisko developera Golang z perspektywy osoby z \"wewnątrz\""
aliases = ["/2017/10/21/w-poszukiwaniu-pracy/"]
+++

## Wstęp

Pierwsza stabilna wersja języka została wydana w marcu 2012 roku.
Od tamtego czasu mineło już ponad 5 lat.

W tym czasie w rankingu [TIOBE](https://www.tiobe.com/tiobe-index/) język Go zanotował wzrost z 0.086% do 1.357% i plasuje się na pozycji 20.
Mogłoby się to wydawać niewiele, ale jak porównamy to do Javascript'u posiadającego obecnie 3% udziału w rynku i będącego na pozycji 6, perspektywa trochę się zmienia.

Język zajmuję też 9 miejsce pod względem ilości otwartych pull requestów na [githubie](https://octoverse.github.com), posiadając 285 tysięcy kontrybucji.
Dla porównania - Java, znajdująca się na miejscu drugim, posiada ich 986 tysięcy.

Szacuje się, że na tą chwilę na świecie jest przynajmniej [500 tysięcy](https://research.swtch.com/gophercount) developerów Go. 
Język otrzymał wsparcie większości edytorów oraz środowisk deweloperskich.
Na etapie Early Access jest [komercyjne IDE](https://www.jetbrains.com/go/) od Jetbrains.
Golang jest także preinstalowany w wielu dystrybucjach linuxa.
 
Dlaczego o tym wszystkim piszę? 
Odnoszę wrażenie, że pomimo tych wszystkich lat __Golang__ jest postrzegany jako coś egzotycznego i niszowego.
Przekłada się to bezpośrednio na podejscie kandydatów do procesu rekrutacji - 
świadomość kandydata stoi w miejscu, a oczekiwania coraz bardziej doświadczonych zespołów rosną.

Tym wpisem chciałbym wpłynąć na świadomość programistów w zakresie wymagań jakie są przed nimi stawiane. Lista poniżej jest moim **subiektywnym** zestawieniem zagadnień, których znajomość w moich oczach pozytywnie wpływa na ocene kandydata.


## Filozofia

__Golang__ nie bez powodu uchodzi za język prosty do nauki. 
Jego składnia jest prosta, liczba słów kluczowych niewielka, a biblioteka standardowa kompletna. 

W mojej ocenie jest to także jezyk (co zaskakujące) znacznie trudniejszy do opanowania niż by się mogło wydawać.

Trywialnym jest napisanie bylejakiego kodu. 
Trochę trudniejsze ale ciągle łatwe jest napisanie kodu przekombinowanego, z masą niepotrzebnej abstrakcji i komponentów.
Napisanie kodu, który jest genialny w swojej prostocie jest naprawde trudne.
A o prostotę przecież tutaj chodzi.

Z moich obserwacji wynika, że czas potrzebny programistom na oduczenie się nawyków przyniesionych z wcześniejszych technologi jest nieakceptowalnie długi.
Nie każdy także musi podzielać wizję twórców Go, a to co przyciągneło go do języka to obietnica wydajności i wsparcia równoległego przetwarzania.
Co do zasady sądze że powinno odrzucać się kandydatów, którzy nie podzielają filozofii __Go__.
Jest to zwyczajnie zbyt kosztowne i niebezpieczne dla kultury wytworzeonej w zespole.

* [Simplicity is Complicated](https://www.youtube.com/watch?v=rFejpH_tAHM)
* [Proverbs](https://go-proverbs.github.io)

## Narzędzia

Jednym z największych atutów __Go__ jest jego środowisko developerskie. 
Poza językiem instalator instaluje również szereg narzędzi, powiedziałbym niezbędnych w codziennej pracy takich jak `cover`, `fmt`, `vet`, `trace` czy `pprof`.
Znajomość tych narzędzi może być przydatna w trakcie rozmowy kwalifikacyjnej podczas sesji pair-programing, kiedy developer poprosi was o optymalizację albo naprawienie danej aplikacji.

Ponadto, przydać się może doświadczenie z bardziej zaawansowanymi opcjami kompilacji. 
Dla przykładu kiedy chcemy przeprowadzić [Escape Analysis](https://en.wikipedia.org/wiki/Escape_analysis).

## Współdzielenie danych i synchronizacja

Kanały są jednym z najbardziej rozpoznawalnych aspektów __Go__. 
Ich poprawne stosowanie przysparza kandydatom jednak wiele problemów. 
To zły znak. 
Przejawia się to najczęściej zbyt dużą ich ilością, dodatkowo niepotrzebnie użytą strukturą `sync.WaitGroup` czy poprostu stosowaniem ich tam, gdzie jest to zbędne.

> Share memory by communicating; don't communicate by sharing memory

Potencjalny kandydat musi rozumieć kiedy stosować kanały, [mutex](https://pl.wikipedia.org/wiki/Problem_wzajemnego_wykluczania) czy też może sprzętową synchronizację dostępną w paczce [sync/atomic](https://golang.org/pkg/sync/atomic/).

* [The Little Book of Semaphores](http://greenteapress.com/wp/semaphores/)
* [Everything You Always Wanted to Know About Synchronization but Were Afraid to Ask](http://sigops.org/sosp/sosp13/papers/p33-david.pdf)
* [Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines)
  
## Przerwania
We wszelkiego rodzaju aplikacjach webowych czy też systemach rozproszonych, możliwość przerwania przetwarzania żądania jest na wagę złota. 
Taki mechanizm pozwala kontrolować zużycie zasobów (ich zwalnianie) oraz przestrzec się przed katastrofalnym w skutkach efektem domina.

Znajomość paczki [context](https://golang.org/pkg/context) jest tutaj kluczowa. 
Za jej pomocą jesteśmy w stanie anulować żądania oraz procesy, a także określać górną granice czasu, w którym nasza logika ma zostać wykonana. 

Temat zdecydowanie zasługuje na osobny wpis.

* [Go Concurrency Patterns: Context](https://blog.golang.org/context)
  

## Whitebox Monitoring/Tracing
Ostatecznie nasz kod ląduje na produkcji. 
Rolą dewelopera jest udostępnienie metryk, które pomogą określić wydajność aplikacji oraz zweryfikować jej poprawne działanie.

Do monitorowania najczęstszym wyborem, w przypadku zespołów, czysto Golangowych jest [Prometheus](https://prometheus.io). 
Najpewniej żadne zadanie testowe nie będzie wymagać jego zastosowania, ale znajomość zagadnienia jest jak najbardziej na plus.

[Tracing](https://en.wikipedia.org/wiki/Tracing_(software)) może zostać zaimplementowany na wiele sposobów. 
Na początek, znajomość paczki [x/net/trace](https://godoc.org/golang.org/x/net/trace) powinna wystarczyć.

## Podsumowanie
Mimo iż __Golang__ jest stosowany przewarznie w zastosowaniach serwerowych/cloudowych, nie musi to być zasadą.

Wiele tematów takich jak testowanie, benchmnarking czy stosowanie paczki `http` umyślnie zignorowałem. 
Skupiłem się jedynie na aspektach, które najczęsciej są pomijane przez kandydatów, których miałem przyjemność ewaluować.

Mam nadzieję, że powyższy tekst pomoże komuś w znalezieniu upragnionej pracy. 
Wszystkim przyszłym [gopherom](https://blog.golang.org/gopher) życzę powodzenia. 



