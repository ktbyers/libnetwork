package libnetwork

type networkNamespace struct {
	path       string
	interfaces []*Interface
}

// Create a new network namespace mounted on the provided path.
func NewNamespace(path string) (Namespace, error) {
	if err := Reexec(ReexecCreateNamespace, path); err != nil {
		return nil, err
	}
	return &networkNamespace{path: path}, nil
}

func (n *networkNamespace) AddInterface(i *Interface) error {
	if err := Reexec(ReexecMoveInterface, i.SrcName, i.DstName); err != nil {
		return err
	}
	n.interfaces = append(n.interfaces, i)
	return nil
}

func (n *networkNamespace) Interfaces() []*Interface {
	return n.interfaces
}

func (n *networkNamespace) Path() string {
	return n.path
}