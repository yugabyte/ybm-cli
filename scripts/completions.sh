#!/bin/sh
# scripts/completions.sh
# From https://carlosbecker.com/posts/golang-completions-cobra/
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run main.go completion "$sh" >"completions/ybm.$sh"
done