package annotations

type AnnotationTarget struct {
	File       string
	StructName string
	MethodName string
	Package    string
}

type AnnotationHandler func(target AnnotationTarget, args map[string]string) error

type AnnotationProcessor struct {
	handlers map[string]AnnotationHandler
}

func NewProcessor() *AnnotationProcessor {
	return &AnnotationProcessor{handlers: make(map[string]AnnotationHandler)}
}

func (p *AnnotationProcessor) Register(name string, handler AnnotationHandler) {
	p.handlers[name] = handler
}

func (p *AnnotationProcessor) Handle(name string, target AnnotationTarget, args map[string]string) error {
	if h, ok := p.handlers[name]; ok {
		return h(target, args)
	}
	return nil
}
