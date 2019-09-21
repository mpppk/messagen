package messagen

type State map[string]Message

func (s State) Set(defType DefinitionType, msg Message) {
	s[string(defType)] = msg
}

func (s State) Get(defType DefinitionType) (Message, bool) {
	v, ok := s[string(defType)]
	return v, ok
}
