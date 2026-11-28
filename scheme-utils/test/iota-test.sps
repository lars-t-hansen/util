;; -*- mode: scheme -*-

(import (scheme base) (scheme write))
(import (test harness))

(import (util iota))

(check (iota 0) '())
(check (iota 10) '(0 1 2 3 4 5 6 7 8 9))

(display "  iota-test OK")
(newline)
