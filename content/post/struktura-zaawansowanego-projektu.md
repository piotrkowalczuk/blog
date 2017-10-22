+++
category = ["post"]
date = "2016-12-22T12:33:03+02:00"
tags = ["golang"]
title = "wersjonowanie kodu napisanego w Go"
draft = true
+++

# Wstęp

Go jako środowisko jest zdeterminowane mocno przez struktórę i organizację pracy wewnątrz Google. 
Gigant ten wytwarza oprogramowanie w oparciu o tzw. [monorepo](http://danluu.com/monorepo/). 
W tym przypadku jakikolwiek problem z zalerznościami nie istnieje. 
Kod całej organizacji znajduje się w jednym miejscu. 

Kiedy Go doczekało się swojej pierwszej [stabilnej wersji](https://blog.golang.org/go-version-1-is-released) na początku 2012 roku temat zalerzności został całkowicie pominięty.
To były ciężkie czasy. 
W zasadzie jedynym rozwiązaniem było posiadanie wielu [przestrzeni roboczych](https://golang.org/doc/code.html#Workspaces) i przesyłanie ich w całości do repozytorium.

Jednym z pierwszych narzędzi był [godep](https://github.com/tools/godep). 
Kardynalnym problemem który mu towarzyszył, była potrzeba przepisywania importów. 
Całe szczęście cały proces był zautomatyzowany.

Wiele się zmieniło w 2015 roku kiedy światło dziennie ujrzała wersja 1.5.
Wprowadziła ona możliwość nadpisywania zależności poprzez umieszczenie ich w katalogu `vendor`. 
Mechanizm działania jest doskonale opisany w tym [dokumencie](https://docs.google.com/document/d/1Bz5-UB7g2uPBdOx-rw5t9MxJwkfpx90cqG9AFL0JAYo/view).

Od tego momentu rozpoczoł się wysyp narzędzi które działały w oparciu o wyrzej wymieniony mechanizm.


# Semantic Versioning

Semantic versioning (aka [semver](http://semver.org)) jest to metoda oznaczania oprogramowania. 
Została ona opracowana przez [Tom'a Preston-Werner'a](http://tom.preston-werner.com) (współzałorzyciela Github.com).
Jest obecnie najszerzej (twierdze tak tylko w oparciu o swoje prywatne doświadczenia) stosowana metoda wsród społeczności Go. 
W durzym uproszczeniu. 
Identyfikator składa się z trzech członów oddzielonych kropkami:`{major}.{minor}.{patch}`.

* `major` - zmiana tego numeru oznacza wprowadzenie zmian publicznego API które nie jest kompatybilne wstecz
* `minor` - zmienia publiczne API, ale kompatybilność wsteczna jest nienaruszona
* `patch` - pomniejsze zmiany

# Praktyczne zastosowanie

Na wstępie warto rozróżnić że inaczej podejdziemy do aplikacji która jest kompilowana do kodu wykonywalnego a inaczej do biblioteki.
Nie mniej jednak w obu przypadkach semver odgrywa tak samo ważną rolę.


Obecnie dostępnych jest wiele menadżerów. 
Bardziej lub mniej kompletną listę możecie znaleźć na oficjalnej [wiki](https://github.com/golang/go/wiki/PackageManagementTools).
Osobiście preferuje [glide](glide.sh) i ten właśnie manager użyję tutaj.

## Biblioteka


## Aplikacja

Na początek warto napisać prostą aplikację której zadaniem będzie wyswietlenie swojej własnej wersji.
```go
package main

import (
	"flag"
	"fmt"
	"os"
)

var version string

type config struct {
	version bool
}

func main() {
	c := config{}
	flag.BoolVar(&c.version, "version", false, "prints version")
	flag.Parse()

	if c.version {
		fmt.Print(version)
		os.Exit(0)
	}

	fmt.Println("program executes here...")
}
```

## Inicjowanie repozytorium

Naszym celem jest utworzenie pierwszego commita który będzie oznaczony stosowną wersją.

```bash
git init
git remote add origin <zewnetrzne-repozytorium>
touch README.md
git add -A
git commit -m "initial commit"
git tag v0.1.0
git push --all
```

## Oznaczanie pliku wykonywalnego

Dobrą praktyką jest opisanie każdego pliku wykonywalnego wersją której ten plik odpowiada.
Warto ten proces zautomatyzować żeby uniknąć.

Pierwszym krokiem będzie utworzenie pliku `Makefile` o zawartości:
```
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS = -X 'main.version=$(VERSION)'

build:
	@go build -ldflags "${LDFLAGS}" -a -o bin/app .
```

Zawiera on tylko jeden target. `build` ma zadanie kompilacje aplikacji oraz wstrzyknięcie wersji za pomocą flagi `-ldflags`.
Wartość domyślna zmiennej `version` w tak skompilowanym programie będzie równa wynikowi komendy `git describe --tags --always --dirty`.



* wstep (google, monorepo)
* obecnie dostepne metody
* przyszłość