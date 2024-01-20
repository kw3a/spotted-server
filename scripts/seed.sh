#!/bin/bash
cd sql/seeders
go build -o seeder
./seeder
