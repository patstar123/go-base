package comm

type Args struct {
	A, B int
}

type Flags struct {
	BoolValue   bool   `yaml:"bool_value"`
	StringValue string `yaml:"string_value"`
	U32Value    uint32 `yaml:"u_32_value"`
	U16Value    uint16 `yaml:"u_16_value"`
}

type Args2 struct {
	StringValue1 string
	StringValue2 string
	Flags1       *Flags
}
