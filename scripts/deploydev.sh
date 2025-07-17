ROOT=/opt/www/MyAPI/dev
TMP=/opt/www/MyAPI/tmp
BRANCH=master
IDENTITY="username@domain.com"
REPO="API REPO"
GZIPPED_TAR="mytarfile.tgz"
dirs_num_to_keep=6

export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/usr/local/go/bin

function message {
  echo "----> $1"
}

function echo_cmd {
  echo "----> $1"
}

function echo_success {
  echo "----> $1"
}

function echo_fail {
  echo "----> Command failed with code $1"
}

function cleanup {
  if [[ -z $rno ]]; then
    echo_cmd "rm -rf $ROOT/releases/$rno"
    rm "-rf" "$ROOT/releases/$rno"
    if [[ $? -eq 0 ]]; then
      echo_success "OK"
    else
      echo_fail "$?"
      cleanup
      exit 1
    fi
  fi
}

function check_result {
  if [[ $1 -eq 0 ]]; then
    echo_success "OK"
  else
    echo_fail "$1"
    cleanup
    exit 1
  fi
}

######## create folders
message " $IDENTITY Create subdirs"
echo_cmd "mkdir -p $ROOT/shared"
mkdir "-p" "$ROOT/shared"
check_result $?

echo_cmd "mkdir -p $ROOT/releases"
mkdir "-p" "$ROOT/releases"
check_result $?

echo_cmd "mkdir -p $ROOT/tmp"
mkdir "-p" "$ROOT/tmp"
check_result $?

echo_cmd "mkdir -p $ROOT/log"
mkdir "-p" "$ROOT/log"
check_result $?

message " $IDENTITY Create shared dirs" 

echo_cmd "mkdir -p $ROOT/shared/assets"
mkdir "-p" "$ROOT/shared/assets"
check_result $?

echo_cmd "mkdir -p $ROOT/shared/node_modules"
mkdir "-p" "$ROOT/shared/node_modules"
check_result $?

echo_cmd "cd $ROOT"
cd "$ROOT"
check_result $?

############ get code
message " $IDENTITY Determine release number"
rno="$(readlink "$ROOT/current")"
rno="$(basename "$rno")"
((rno = $rno + 1))

message " $IDENTITY Clone code into release $rno"
# skip - extract uploaded tar instead
echo_cmd "cd $ROOT"
cd "$ROOT"
check_result $?

#echo_cmd "git clone -b $BRANCH $REPO $ROOT/releases/$rno"
rm -rf $ROOT/releases/$rno
mkdir -p $ROOT/releases/$rno
tar xzf $GZIPPED_TAR -C $ROOT/releases/$rno

#git "clone" "-b" "$BRANCH" "$REPO" "$ROOT/releases/$rno"

check_result $?

################ create links
message "$IDENTITY Link shared dirs"
echo_cmd "cd $ROOT/releases/$rno"
cd "$ROOT/releases/$rno"
check_result $?

echo_cmd "mkdir -p ."
mkdir "-p" "."
check_result $?
 

# cd "$ROOT/releases/$rno/public"
# [ -h dist ] && unlink dist

# echo_cmd "ln -s $ROOT/shared/dist dist"
# ln -s $ROOT/shared/dist dist
# check_result $?

cd "$ROOT/releases/$rno"

echo_cmd "copy files and start script"

mkdir -p bin public/files
cp $ROOT/shared/startProduction.sh bin/start.sh
cp $ROOT/shared/wait-for-it.sh bin/wait-for-it.sh

chmod 775 bin/start.sh

################## run scripts
message " $IDENTITY Run pre-start scripts"
echo_cmd "make compile"
make compile 
check_result $?

#rm -rf .git api src Dockerfile makefile go.mod go.sum main.go .github 

message " $IDENTITY Update current link"
echo_cmd "cd $ROOT"
cd "$ROOT"
check_result $?

message "This is the point of no return, any errors from hereon and you need to cleanup manually"
if [[ -L current ]]; then
  echo_cmd "rm current"
  rm "current"
  if [[ $? -eq 0 ]]; then
    echo_success "OK"
  else
    echo_fail "$1"
    echo_fail "previous link was not removed, no further action taken"
    exit 1
  fi
fi

echo_cmd "ln -s releases/$rno current"
ln "-s" "releases/$rno" "current"
if [[ $? -eq 0 ]]; then
  echo_success "OK"
else
  echo_fail "$1"
  echo_fail "This is really bad news.  We couldn't update the current->release link so your site is broken."
  exit 1
fi

#message " $IDENTITY do post install stuff e.g. restart service"


######### cleanup old releases
message " $IDENTITY Cleaning release dir"
echo_cmd "cd $ROOT/releases"
cd "$ROOT/releases"

release_dirs="$(find "." "-maxdepth" "1" "-mindepth" "1" "-type" "d" "-printf" "%f\n")"
num_dirs="$(echo_cmd "$release_dirs" | wc -l)"

if ((num_dirs > dirs_num_to_keep)); then
  ( 
    ((dirs_num_to_remove = $num_dirs - $dirs_num_to_keep))
    echo "$release_dirs" | sort -n | head -n$dirs_num_to_remove
  ) | (
    while read rm_dir; do
      echo_cmd "rm -rf $rm_dir"
      rm "-rf" "$rm_dir"

    done
  )
fi
