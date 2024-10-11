#!/bin/bash

set -e

usage() {
  echo 'Usage: $1 output_dir binary_path ...'
}

output_dir=$1
if [[ -e "$output_dir" ]]; then
  echo "$output_dir exists!" >&2
  exit 1
fi
mkdir -p $output_dir

shift
binary_paths=$@

if [[ ${#binary_paths[@]} == 0 ]]
then
  usage
  exit 1
fi

echo "Copying dependent libs to $output_dir"

find_deps() {
  local paths=$@
  find $paths -type f -o -type l -and -executable -or -name "*.so*" | \
    xargs ldd | \
    grep -o '/.*' | awk '{print $1}' | grep -v ':$' | sort -u
}

copy_lib_and_links() {
  local lib=$1
  local output_dir=$2

  while [ 0 ]; do
    file=$(basename $lib)
    dir=$(dirname $lib)
    mkdir -p ${output_dir}/$dir
    cp -a $lib ${output_dir}/$dir
    lib=$(readlink "${dir}/$file" || true)
    if [[ -z "$lib" ]]; then
      break
    fi
    libdir=$(dirname $lib)
    if [[ "${libdir:0:1}" != "/" ]]; then
      lib=${dir}/$lib
    fi
  done
}

libs="$(find_deps $binary_paths)"
transitive_libs="$(find_deps $libs)"

for lib in $libs $transitive_libs
do
  if [[ ! -e ${output_dir}/$lib ]]; then
    copy_lib_and_links $lib $output_dir
    echo "Pulled in $lib"
  fi
done

echo "Processed $(wc -w <<< $libs) libraries"

echo "Checking for missing lib dependencies..."

# Look for all of the deps now in the output_dir and make sure we have them
new_deps=$(find_deps $output_dir)
for dep in $new_deps
do
  stat ${dep} >/dev/null
  if [[ $? != 0 ]]; then
    echo "Missing dependency in target dir: $dep" >&2
    exit 1
  fi
done

echo "Everything is there!"
