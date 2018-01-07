.PHONY : build clean run
comp = go
main = main.go intVec.go

run : $(main)
	$(comp) run $(main)

build : $(main)
	$(comp) build -o zt $(main)

clean : 
	-@rm zt
	-@rm log
