(define-library (util string-find)
  (export string-find)
  (import (scheme base))
  (begin

    ;; Returns index or #f
    (define (string-find s c)
      (let loop ((i 0))
        (cond ((= i (string-length s)) #f)
              ((char=? (string-ref s i) c) i)
              (else (loop (+ i 1))))))))
