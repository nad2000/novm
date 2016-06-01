package novm

import (
	"crypto/sha1"
	"os"
)

func cleanup()
...

func packdir()
...

func unpackdir()
...

func libexec()
...

func asbool()
...

func clear_cloexec()
...

func copy(dst string, src string, sparse bool, hash bool) string {
// in the python code both sparse and hash has default value false , need to do that explicitely in go 

  src_1, _ := os.Open(src )
  dst_1, _ := os.Create(dst)
  if hash {
    sha1_1 := sha1.New()
  }

  if sparse {
    # Grab our starting point.
    reloff := os. 
    blah blah
  }

  for {
   chunk := 65536
   if sparse {

     }
   data_1, err := src_1.Read(b[chunk])
   if err != nil {
    break
   }
   dst_1.Write(b)
   if sparse {
       reloff += data_1

   if hash {
      sha1_1....(b)
   }
  }
  dst_1.flush()
  if hash {
    return sha1...()
  }

}





















