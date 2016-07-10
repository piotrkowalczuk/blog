GIT_DIR=../piotrkowalczuk.github.io/.git/
GIT_WORK_TREE=../piotrkowalczuk.github.io/

deploy:
	hugo
	cp -r public/ ../piotrkowalczuk.github.io/
	git add -A
	git commit -m "blog deployment"
	git push