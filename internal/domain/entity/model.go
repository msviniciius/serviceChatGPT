package entity

// como se fosse uma clase se estivessimos usando orientação objeto
// vai servir para guardar info sobre a versão do GPT e quant de tokens
type Model struct {
	Name      string
	MaxTokens int
}

// função contrutora, ela vai me permitir fazer uma nova struct
// * é um ponteiro, entrega a referencia da memoria onde esse dado foi criado
// ele não terá id pois funcionarar como um objeto de valor, não como entidade
func NewModel(name string, maxTokens int) *Model {
	return &Model{
		Name:      name,
		MaxTokens: maxTokens,
	}
}

// m serve para acessar o model
func (m *Model) GetMaxTokens() int {
	return m.MaxTokens
}

func (m *Model) GetModelName() string {
	return m.Name
}
