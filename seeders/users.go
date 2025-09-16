package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/seeders/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func (cfg *SeedersConfig) seedUsers() ([]string, error) {
	IDs := []string{}
	users := []struct {
		Nick        string
		Name        string
		Email       string
		Password    string
		Number      string
		Description string
	}{
		{"jlopez", "Juan Lopez", "jlopez@example.com", "12345juan", "+54 11 3456 7890", "Desarrollador backend especializado en Go y microservicios."},
		{"mtorres", "Maria Torres", "mtorres@example.com", "mariatorres22", "+591 76543210", "Ingeniera de software con experiencia en análisis de datos y ciencia de datos aplicada al sector financiero."},
		{"rgomez", "Ricardo Gomez", "rgomez@example.com", "ricky2025", "+57 320 456 7890", "Administrador de sistemas con foco en seguridad informática y gestión de servidores Linux."},
		{"aflores", "Andrea Flores", "aflores@example.com", "andreaflores77", "+56 9 8765 4321", "Diseñadora UX/UI orientada a optimizar la experiencia del usuario en aplicaciones móviles."},
		//A
		{"pcaceres", "Pablo Caceres", "pcaceres@example.com", "pabloc123", "+598 94 345 678", "Especialista en recursos humanos, con experiencia en reclutamiento de perfiles tecnológicos."},
		{"sluna", "Sofia Luna", "sluna@example.com", "sofiaL!89", "+55 21 91234 5678", "Ingeniera de software fullstack con conocimiento en React, Node.js y arquitecturas en la nube."},
		//B
		{"dcastro", "Daniel Castro", "dcastro@example.com", "danielc2025", "+51 987 654 321", "Scrum Master con amplia experiencia en metodologías ágiles y liderazgo de equipos distribuidos."},
		{"vmendez", "Valeria Mendez", "vmendez@example.com", "valemz12", "+593 99 123 4567", "Analista QA enfocada en pruebas automatizadas y control de calidad de software."},
		{"jramirez", "Jorge Ramirez", "jramirez@example.com", "jorgeR2024", "+52 55 8765 4321", "Arquitecto de software con más de diez años de experiencia en sistemas distribuidos."},
		{"afernandez", "Ana Fernandez", "afernandez@example.com", "anaF!2024", "+54 261 456 7890", "Especialista en ciberseguridad, con enfoque en análisis de vulnerabilidades."},
		{"csilva", "Carlos Silva", "csilva@example.com", "carlitos99", "+55 11 92345 6789", "DevOps Engineer con experiencia en CI/CD y automatización de infraestructura en AWS."},
		{"lreyes", "Laura Reyes", "lreyes@example.com", "laurita2025", "+57 315 234 5678", "Ingeniera en machine learning aplicada a procesamiento de lenguaje natural."},
		{"mrojas", "Marcos Rojas", "mrojas@example.com", "marcoR0j4s", "+591 70123456", "Técnico en soporte TI con experiencia en atención a usuarios y mantenimiento de redes."},
		//C
		{"acampos", "Adriana Campos", "acampos@example.com", "adriC2024", "+56 9 7654 1234", "Reclutadora IT especializada en perfiles de desarrollo web y mobile."},
		{"farias", "Fernando Arias", "farias@example.com", "ferarias25", "+51 944 567 890", "Ingeniero de datos experto en pipelines de ETL y procesamiento en tiempo real."},
		//D
		{"mgutierrez", "Monica Gutierrez", "mgutierrez@example.com", "moniGT2025", "+593 98 765 4321", "Consultora en recursos humanos con experiencia en planes de capacitación tecnológica."},
		{"jvaldez", "Javier Valdez", "jvaldez@example.com", "javiV123", "+55 31 92345 6789", "Ingeniero de software móvil especializado en desarrollo nativo para Android."},
		{"ysalazar", "Yesenia Salazar", "ysalazar@example.com", "yesal2024", "+54 299 456 1234", "Diseñadora gráfica orientada a branding y marketing digital."},
		//E
		{"ogarcia", "Oscar Garcia", "ogarcia@example.com", "oscarg!2025", "+57 300 987 6543", "Project Manager certificado PMP con trayectoria en proyectos de software bancario."},
		{"nmartinez", "Natalia Martinez", "nmartinez@example.com", "nataliam2025", "+598 92 123 456", "Especialista en e-learning y diseño instruccional para capacitaciones corporativas."},
		{"hfuentes", "Hector Fuentes", "hfuentes@example.com", "hect0rF2025", "+51 955 123 456", "Analista de seguridad con experiencia en auditorías de cumplimiento normativo."},
		//F
		{"acarrillo", "Alicia Carrillo", "acarrillo@example.com", "alicarr2025", "+591 78965432", "Coordinadora de recursos humanos enfocada en clima laboral y retención de talento."},
		{"dmarin", "Diego Marin", "dmarin@example.com", "diegomar!n", "+56 9 3456 7890", "Programador frontend con experiencia en Vue.js y aplicaciones progresivas."},
		//G
		{"pmendoza", "Patricia Mendoza", "pmendoza@example.com", "patymdz2025", "+54 351 765 4321", "Consultora en gestión del talento con enfoque en diversidad e inclusión."},
		{"eortega", "Esteban Ortega", "eortega@example.com", "estebano25", "+57 312 654 9870", "Especialista en bases de datos relacionales y NoSQL."},
		{"lgonzalez", "Lucia Gonzalez", "lgonzalez@example.com", "lucy2025", "+55 41 99876 5432", "Ingeniera en inteligencia artificial enfocada en visión por computadora."},
		{"fperalta", "Felipe Peralta", "fperalta@example.com", "felipeP123", "+593 97 654 3210", "Desarrollador de videojuegos con experiencia en Unity y Unreal Engine."},
		//H
		{"lrodriguez", "Laura Rodriguez", "lrodriguez@example.com", "lauRod2024", "+598 93 876 543", "Psicóloga organizacional especializada en evaluación de competencias."},
		{"emartin", "Eduardo Martin", "emartin@example.com", "eduM@rtin25", "+54 341 678 9012", "Ingeniero de software con experiencia en arquitecturas serverless."},
		{"gcarrasco", "Gabriela Carrasco", "gcarrasco@example.com", "gabyc2025", "+51 987 321 654", "Recruiter IT con experiencia en selección de perfiles senior y ejecutivos."},
	}
	for _, u := range users {
		id := uuid.New().String()
		pass, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		err = cfg.DB.SeedUser(cfg.Ctx, database.SeedUserParams{
			ID:          id,
			Nick:        u.Nick,
			Name:        u.Name,
			Email:       u.Email,
			Number:      u.Number,
			Password:    string(pass),
			Description: u.Description,
		})
		if err != nil {
			return nil, err
		}
		IDs = append(IDs, id)
	}
	fmt.Println("Users seeded successfully")
	if err := SaveIDsToFile(IDs, "to_delete/user.txt"); err != nil {
		return nil, err
	}
	return IDs, nil
}
