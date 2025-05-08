package shared

const (
	ErrUserTaken = "El correo ya está siendo utilizado"
	ErrNotMatch       = "El correo y la contraseña no coinciden"
	ErrNotRegistered  = "Este usuario no existe"
	MsgSaved = "Guardado"
)

type Alert struct {
	Ok  bool
	Msg string
}
