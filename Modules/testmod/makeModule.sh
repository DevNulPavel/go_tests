#! /usr/bin/env bash

# Инициализация модуля
# https://habr.com/ru/post/421411/
go mod init github.com/DevNulPavel/GoTests/Module/testmod

git add .
git commit -a -m "Module commit" 
git push
git tag v1.0.0
git push --tags


