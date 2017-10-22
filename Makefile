
GIT_OPTS=--git-dir=../piotrkowalczuk.github.io/.git/ --work-tree=../piotrkowalczuk.github.io/

deploy:
	@hugo
	@cp -r public/ ../piotrkowalczuk.github.io/
	@cp -r examples/ ../piotrkowalczuk.github.io/examples
	@git ${GIT_OPTS} add -A
	@git ${GIT_OPTS} commit -m "blog deployment"
	@git ${GIT_OPTS} push