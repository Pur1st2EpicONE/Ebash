#!/bin/bash

commands=(
    "cd .. && ls"
    "pwd"
    "echo Aboba Testing Inc."
    "ls"
    "cd / && pwd"
    "cd ~ && pwd"
    "ls | wc -l"
    "false && echo no || echo yes"
    "echo \$HOME"
    "echo \$PATH"
    "uname -a"
    "date"
    "whoami"
    "ls -l | head -n 5"
    "true && echo OK | cat"
    "false && echo NOT OK | cat"
    "false || echo OK | cat"
    "true || echo NOT OK | cat"
    "true && echo First && echo Second | cat"
    "false || echo One || echo Two | cat"
    "cat < go.sum | grep 2.4 | sort -r"
    "cat < Makefile | grep .go | sort -r > tmp1.txt && echo OK || echo NOT OK"
    "cd ~ && ls -la | head -3 && echo OK || echo NOT OK"
    "ls | sort | grep Makefile"
    "false && echo NOT OK || echo OK | cat"
    "echo qwe > tmp2.txt && cat tmp2.txt"
)

log=$(mktemp)

for cmd in "${commands[@]}"; do
    tmp1=$(mktemp)
    tmp2=$(mktemp)

    echo "$cmd" | ./ebash 2>&1 > "$tmp1"
    bash -c "$cmd" 2>&1 > "$tmp2"

    if diff -u "$tmp1" "$tmp2" > /dev/null; then
        echo "Test passed: $cmd" | tee -a "$log"
    else
        echo "Test failed: $cmd" | tee -a "$log"
        diff -u "$tmp1" "$tmp2" | tee -a "$log"
    fi

    rm -f "$tmp1" "$tmp2"
done

echo a > test && echo b >> test
echo "echo a > test2" | ./ebash
echo "echo b >> test2" | ./ebash

if diff -u "test" "test2" > /dev/null; then
    echo "Test passed: echo >>" | tee -a "$log"
else
    echo "Test failed: echo >>" | tee -a "$log"
    diff -u "test" "test2" | tee -a "$log"
fi

rm -f "tmp1.txt" "tmp2.txt" ">" "test" "test2"
rm -rf temp_test_dir

if grep -q "Test failed" "$log"; then
    exit 1
fi

exit 0
