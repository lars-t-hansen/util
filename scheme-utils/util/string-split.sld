(define-library (util string-split)
  (export string-split)
  (import (scheme base))
  (begin

    (define (string-split s c)
      (if (string=? s "")
	  (list s)
	  (let loop ((i 0) (start 0) (ss '()))
	    (cond ((= i (string-length s))
		   (reverse (cons (substring s start i) ss)))
		  ((char=? (string-ref s i) c)
		   (loop (+ i 1) (+ i 1) (cons (substring s start i) ss)))
		  (else
		   (loop (+ i 1) start ss))))))))
