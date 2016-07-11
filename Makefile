
deploy:
	@export GIT_WORK_TREE=../piotrkowalczuk.github.io/
	@export GIT_DIR=../piotrkowalczuk.github.io/.git/
	@hugo
	@cp -r public/ ../piotrkowalczuk.github.io/
	@cp -r examples/ ../piotrkowalczuk.github.io/examples
	@git add -A
	@git commit -m "blog deployment"
	@git push