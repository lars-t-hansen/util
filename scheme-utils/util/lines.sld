(define-library (util lines)
  (export lines)
  (import (scheme base))
  (begin

    (define (lines . rest)
      (let ((inp (if (null? rest) (current-input-port) (car rest))))
	(let loop ((ls '()))
	  (let ((l (read-line inp)))
	    (if (eof-object? l)
		(reverse ls)
		(loop (cons l ls)))))))))
