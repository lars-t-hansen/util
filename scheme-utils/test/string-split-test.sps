;; -*- mode: scheme -*-

(import (scheme base) (scheme write))
(import (util string-split))
(import (test harness))

(check (string-split "" #\-) '(""))
(check (string-split ",a," #\,) '("" "a" ""))
(check (string-split "a,b,c," #\,) '("a" "b" "c" ""))
(display "OK")
(newline)
