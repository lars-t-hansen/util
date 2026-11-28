(define-library (util iota)
  (export iota)
  (import (scheme base))
  (begin

    (define (iota n)
      (let loop ((i n) (r '()))
        (if (= i 0)
            r
            (loop (- i 1) (cons (- i 1) r)))))))
