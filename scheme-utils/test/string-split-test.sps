;; -*- mode: scheme -*-

(import (scheme base) (scheme write))
(import (test harness))

(import (util string-split))

(check (string-split "" #\-) '(""))
(check (string-split ",a," #\,) '("" "a" ""))
(check (string-split "a,b,c," #\,) '("a" "b" "c" ""))

(display "  string-split-test OK")
(newline)
