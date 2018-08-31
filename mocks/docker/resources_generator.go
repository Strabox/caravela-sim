package docker

type ResourcesGenerator interface {
	Generate() (int, int, int)
}
