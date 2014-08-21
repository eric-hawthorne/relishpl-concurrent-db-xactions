// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

// This "general os" package provides platform-independent wrapper versions 
// of common filesystem functions. These are some of the functions found in platform-specific form in
// Go's os and io/ioutil packages.
//
// Methods of this package which accept a path argument accept the "/" separated (unix-style) filesystem paths,
// and translate as needed to Windows filesystem paths.
// Similarly, if a method in gos package returns a path, it will always be the "/" separated (unix-style) path.

package gos

/*
   file_util.go - convenience methods for multi-platform filesystem operations.
*/

import (
  "path/filepath"
  "os"
  "io/ioutil"
  )


func Mkdir(name string, perm os.FileMode) error {
   return os.Mkdir(filepath.FromSlash(name), perm)
}

func MkdirAll(name string, perm os.FileMode) error {
   return os.MkdirAll(filepath.FromSlash(name), perm)
}

func Create(name string) (file *os.File, err error) {
  return os.Create(filepath.FromSlash(name))
}

func Open(name string) (file *os.File, err error) {
  return os.Open(filepath.FromSlash(name)) 
}

func OpenFile(name string, flag int, perm os.FileMode) (file *os.File, err error) {
   return os.OpenFile(filepath.FromSlash(name), flag, perm)
}

func Lstat(name string) (os.FileInfo, error) {
   return os.Lstat(filepath.FromSlash(name))
}

func Stat(name string) (os.FileInfo, error) {
   return os.Stat(filepath.FromSlash(name))
}

func Getwd() (pwd string, err error) {
   pwd, err = os.Getwd()
   return filepath.ToSlash(pwd),err
}

func Remove(name string) error {
   return os.Remove(filepath.FromSlash(name))
}

func RemoveAll(path string) error {
   return os.RemoveAll(filepath.FromSlash(path))
}

func Rename(oldname, newname string) error {
   return os.Rename(filepath.FromSlash(oldname), filepath.FromSlash(newname))
}

func TempDir() string {
  return filepath.ToSlash(os.TempDir())
}


func WriteFile(filename string, data []byte, perm os.FileMode) error {
  return ioutil.WriteFile(filepath.FromSlash(filename), data, perm)
}

func ReadFile(filename string) ([]byte, error) {
   return ioutil.ReadFile(filepath.FromSlash(filename))
}

func ToOsSpecificPath(filePath string) string {
   return filepath.FromSlash(filePath))
}