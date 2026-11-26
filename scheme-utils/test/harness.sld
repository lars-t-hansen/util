(define-library (test harness)
  (export check)
  (import (scheme base))
  (begin

    (define (check got expected)
      (if (not (equal? got expected))
	  (error "Got " got " but expected " expected)))))

