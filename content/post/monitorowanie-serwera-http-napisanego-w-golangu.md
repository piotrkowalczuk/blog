+++
category = ["przykłady"]
date = "2017-11-14T00:12:21+01:00"
lastmod = "2017-11-19T00:12:51+01:00"
tags = ["golang","prometheus","monitoring", "instrumentacja"]
title = "Monitorowanie serwera HTTP napisanego w Golangu"
description = "Tutorial pokazujący jak zaimplementować instrumentację serwera HTTP przy pomocy Prometheus'a"
aliases = [
    "/2017/11/14/prometheus-monitorowanie-serwera-http/",
    "/blog/monitorowanie-serwera-http-napisanego-w-golangu/"
]
+++

## Czym jest Prometheus?
[Prometheus](https://prometheus.io) jest to ekosystem do monitorowania napisany przez programistów z [SoundCloud](https://soundcloud.com). 
Jak możecie się przekonać przeglądając oficjalne konto na [githubie](https://github.com/prometheus), większość środowiska jest napisana w [Go](http://golang.org).
Od 2016 roku projekt jest też częścią [Cloud Native Computing Foundation](https://www.cncf.io) obok takich rozwiązań jak [kubernetes](https://kubernetes.io), [gRPC](http://grpc.io) czy [OpenTracing](http://opentracing.io).
Daje nam to pewność, że projekt będzie rozwijany przez długie lata, będzie ewoluował razem z resztą środowiska, a także wsparcie dla __Go__ będzie stało na najwyższym poziomie.

Skupię się tutaj na ostatniej wersji, oznaczonej tagiem `v0.8.0`. 
W tej wersji wiele funkcji zostało oznaczonych jako `DEPRECATED` i zostaną one przeze mnie pominięte.
Zalecana wersja Go to 1.9+.

## Biblioteka

Zasadniczo __Prometheus__ jako serwer centralny musi być świadomy istnienia aplikacji która jest monitorowana. 
Tylko wtedy jest on w stanie pobrać metryki ze wskazanego endpointu.
Z pomocą przychodzi nam biblioteka [promhttp](https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp), która jest częścią składową oficjalnej [paczki](https://github.com/prometheus/client_golang/tree/master/prometheus). 

[promhttp.HandlerFor](https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#HandlerFor) pozwala utworzyć endpoint dla danego [prometheus.Gatherer'a](https://godoc.org/github.com/prometheus/client_golang/prometheus#Gatherer). 
Interfejs ten jest na przykład implementowany przez [prometheus.DefaultRegisterer](https://godoc.org/github.com/prometheus/client_golang/prometheus#pkg-variables).

Ponadto biblioteka ta zawiera garść dekoratorów które, pozwolą nam zbierać informacje na temat naszej aplikacji:

* [promhttp.InstrumentHandlerCounter](https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#InstrumentHandlerCounter) - całkowita liczba przetworzonych żądań
* [promhttp.InstrumentHandlerDuration](https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#InstrumentHandlerDuration) - czas trwania żądania
* [promhttp.InstrumentHandlerTimeToWriteHeader](https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#InstrumentHandlerTimeToWriteHeader) - podobnie jak poprzedni tylko do czasu wysłania nagłówków
* [promhttp.InstrumentHandlerInFlight](https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#InstrumentHandlerInFlight) - liczba obecnie przetwarzanych żądań (w trakcie)
* [promhttp.InstrumentHandlerRequestSize](https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#InstrumentHandlerRequestSize) - wielkość żądania

## Implementacja

Jak widać, żeby zacząć nie trzeba się wiele napracować. Większość potrzebnych nam składników jest już dostępna.
Brakujący element to [kolektory](https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector) które musimy zainicjować własnoręcznie.

```go
duration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Namespace: "acme",
        Subsystem: "your_app",
        Name:      "http_durations_histogram_seconds",
        Help:      "Request time duration.",
    },
    []string{"code", "method"},
)
requests := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Namespace: "acme",
        Subsystem: "your_app",
        Name:      "http_requests_total",
        Help:      "Total number of requests received.",
    },
    []string{"code", "method"},
)
```

Oba one muszą zostać zarejestrowane, a następnie przekazane jako argument do wyżej wymienionych dekoratorów.
Możemy trochę usprawnić ten proces poprzez wprowadzenie dodatkowej struktury. 

```go
type decorator struct {
	duration *prometheus.HistogramVec
	requests *prometheus.CounterVec
}
```

Aby spełniać swoje zadanie, struktura ta powinna implementować interfejs [prometheus.Collector](https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector).

```go
// Describe implements prometheus Collector interface.
func (d *decorator) Describe(in chan<- *prometheus.Desc) {
	d.duration.Describe(in)
	d.requests.Describe(in)
}

// Collect implements prometheus Collector interface.
func (d *decorator) Collect(in chan<- prometheus.Metric) {
	d.duration.Collect(in)
	d.requests.Collect(in)
}
```

Dodatkowo, możemy zredukować duplikację kodu implementując dodatkową metodę.
Jej zadaniem będzie dekorowanie danego handlera szeregiem funkcji.

```go
func (d *decorator) instrument(handler http.Handler) http.Handler {
	return promhttp.InstrumentHandlerDuration(
		d.duration,
		promhttp.InstrumentHandlerCounter(
			d.requests,
			handler,
		),
	)
}
```

Naszym ostatnim krokiem będzie połączenie wszystkiego ze sobą i udostępnienie metryk. 

```go
func main() {
    dec := &decorator{
        // inicjalizacja
    }

    prometheus.DefaultRegisterer.Register(dec)

    go func() {
        dbg := http.NewServeMux()
        dbg.Handle("/metrics", promhttp.HandlerFor(
            prometheus.DefaultGatherer,
            promhttp.HandlerOpts{},
        ))
        http.ListenAndServe("0.0.0.0:8081", dbg)
    }()

    app := http.NewServeMux()
    app.Handle("/", dec.instrument(&handler{}))
    http.ListenAndServe("0.0.0.0:8080", app)
}
```

Aplikacja przez nas napisana będzie nasłuchiwać na dwóch portach. 
Pierwszy `8080`, zarezerowany dla aplikacji właściwej. 
Drugi `8081`, na którym prometheus będzie miał dostęp do metryk.
Chciałbym zwrócić uwagę, że router został w drugim przypadku zastosowany nie bez powodu.
Pozwoli on w przyszłości udostępnić na tym samym porcie healthcheck, czy też endpointy [pprof](https://golang.org/pkg/net/http/pprof/).

## Weryfikacja

Aby sprawdzić, czy nasze demo zwraca poprawny wynik, posłużymy się aplikacją powłoki systemowej `curl`.

```
$ curl http://localhost:8080
It works!
$ curl -s localhost:8081/metrics | grep 'acme_your_app_http_requests_total{code="200",method="get"}'
acme_your_app_http_requests_total{code="200",method="get"} 1
$ curl http://localhost:8080
It works!
$ curl -s localhost:8081/metrics | grep 'acme_your_app_http_requests_total{code="200",method="get"}'
acme_your_app_http_requests_total{code="200",method="get"} 2
```

Zwracana wartość odpowiada ilości wysłanych żądań. 
Monitoring działa bez zarzutu. 


## Podsumowanie

Aby utrzymać przejrzystość, przykłady nie zawierają wszystkich wspieranych metryk. 
Podczas ich implementacji warto zapoznać się z dokumentacją dekoratorów. 
Znajdują się tam informacje o wspieranych etykietach.

Pełny kod źródłowy aplikacji można znaleźć [tutaj](https://github.com/piotrkowalczuk/blog/tree/master/examples/prometheus-monitorowanie-serwera-http).