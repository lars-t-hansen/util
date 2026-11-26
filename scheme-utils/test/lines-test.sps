;; -*- mode: scheme -*-

(import (scheme base) (scheme write) (scheme file))
(import (test harness))

(import (util lines))

(let ((inp (open-input-file "/dev/null")))
  (check (lines inp) '()))

(let ((inp (open-input-file "lines-test.txt")))
  (check (lines inp) '("hi" "there")))

(with-input-from-file "lines-test.txt"
  (lambda ()
    (check (lines) '("hi" "there"))))

(display "  lines-test OK")
(newline)
