;; -*- mode: scheme -*-

(import (scheme base) (scheme write))
(import (test harness))

(import (util string-find))

(check (string-find "" #\-) #f)
(check (string-find "a" #\-) #f)
(check (string-find "abcde" #\a) 0)
(check (string-find "abcde" #\c) 2)
(check (string-find "abcde" #\e) 4)

(display "  string-find-test OK")
(newline)
