build: 
    go build
    rm -rf dist
    mkdir dist
    mv vreemdebedoening ./dist
    cp -r ./templates ./dist/templates
    cp -r ./public ./dist/public

deploy:
    just build
    tar -czf vreemdebedoening.tar.gz dist
    rm -rf dist

run:
    go run .