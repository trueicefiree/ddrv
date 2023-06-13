package ddrvfs

type Fs interface {
    Get(id string, isDir bool) (*File, error)
    GetChild(id string) ([]*File, error)
    Create(name, parent string, isDir bool) (*File, error)
    Update(id, parent, name string, isDir bool) (*File, error)
    Delete(id string) error
    GetFileNodes(id string) ([]*Node, error)
    CreateFileNodes(id string, nodes []*Node) error
}
