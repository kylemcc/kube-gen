# v0.4.0

**Changes:**

* Add template functions:
    * `add` - add two `int`s
    * `pathJoinSlice` - wrapper around filepath.Join that accepts a slice
    * `slice` - implement the slice operator (`{{ slice $s 0 3 }}` is equivalent to `s[0:3]`)


# v0.3.0

**Changes:**

* Managed dependencies with [dep](https://github.com/golang/dep)

* Support authentication via `-kubeconfig` flag

    * https://github.com/kylemcc/kube-gen/pull/1
