#!/bin/bash

if [ -f .env ]; then
    source .env
fi

cd sql/schema
goose mysql $DATABASE_URL down-to 0
cd ../triggers
goose mysql $DATABASE_URL down-to 0