package resources

import "fmt"

// Resources represent of the resources that a user can ask for a container to have available.
type Resources struct {
	cpus int
	ram  int
}

// NewResources creates a new resource combination object.
func NewResources(cpus int, ram int) *Resources {
	return &Resources{cpus: cpus, ram: ram}
}

// AddCPUs adds a given number of cpus to the resources.
func (r *Resources) AddCPUs(cpus int) {
	r.cpus += cpus
}

// AddRAM adds a given amount of ram to the resources.
func (r *Resources) AddRAM(ram int) {
	r.ram += ram
}

// Add adds a given combination of resources to the receiver.
func (r *Resources) Add(resources Resources) {
	r.cpus += resources.CPUs()
	r.ram += resources.RAM()
}

// Sub subtracts a given combination of resources to the receiver.
func (r *Resources) Sub(resources Resources) {
	r.cpus -= resources.CPUs()
	r.ram -= resources.RAM()
}

// SetZero sets the resources to zero.
func (r *Resources) SetZero() {
	r.cpus = 0
	r.ram = 0
}

// SetTo sets the resources into a specific combination of resources.
func (r *Resources) SetTo(resources Resources) {
	r.cpus = resources.CPUs()
	r.ram = resources.RAM()
}

// IsZero returns true if the resources are zero, false otherwise.
func (r *Resources) IsZero() bool {
	return r.cpus == 0 && r.ram == 0
}

// IsValid return true if all resources are greater than 0.
func (r *Resources) IsValid() bool {
	return r.cpus > 0 && r.ram > 0
}

// Contains returns true if the given resources are contained inside the receiver.
func (r *Resources) Contains(contained Resources) bool {
	return r.cpus >= contained.CPUs() && r.ram >= contained.RAM()
}

// Equals returns true if the given resource combination is equal to the receiver.
func (r *Resources) Equals(resources Resources) bool {
	return r.cpus == resources.cpus && r.ram == resources.ram
}

// Copy returns a object that is a exact copy of the receiver.
func (r *Resources) Copy() *Resources {
	res := &Resources{}
	res.cpus = r.cpus
	res.ram = r.ram
	return res
}

// String stringify the receiver resources object.
func (r *Resources) String() string {
	return fmt.Sprintf("Resources: <%d;%d>", r.cpus, r.ram)
}

// CPUs getter.
func (r *Resources) CPUs() int {
	return r.cpus
}

// RAM getter.
func (r *Resources) RAM() int {
	return r.ram
}

// SetCPUs CPUs setter.
func (r *Resources) SetCPUs(cpu int) {
	r.cpus = cpu
}

// SetRAM RAM setter.
func (r *Resources) SetRAM(ram int) {
	r.ram = ram
}
