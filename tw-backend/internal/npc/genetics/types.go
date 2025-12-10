package genetics

// Gene represents a single genetic trait
type Gene struct {
	TraitName   string `json:"trait_name"`
	Allele1     string `json:"allele1"`      // From parent 1
	Allele2     string `json:"allele2"`      // From parent 2
	IsDominant1 bool   `json:"is_dominant1"` // Derived from Allele1 case
	IsDominant2 bool   `json:"is_dominant2"` // Derived from Allele2 case
	Phenotype   string `json:"phenotype"`    // Expressed trait
}

// DNA represents the complete genetic profile
type DNA struct {
	Genes map[string]Gene `json:"genes"`
}

// NewDNA creates an empty DNA profile
func NewDNA() DNA {
	return DNA{
		Genes: make(map[string]Gene),
	}
}

// Common Gene Names
const (
	GeneStrength   = "strength"
	GeneMuscle     = "muscle"
	GeneReflex     = "reflex"
	GeneCoord      = "coordination"
	GeneStamina    = "stamina"
	GeneResilience = "resilience"
	GeneHealth     = "health"
	GeneRecovery   = "recovery"
	GeneCognition  = "cognition"
	GeneLearning   = "learning"
	GenePerception = "perception"
	GeneAnalysis   = "analysis"
	GeneVision     = "vision"
	GeneColor      = "color"
	GeneAuditory   = "auditory"
	GeneRange      = "range"
	GeneHeight     = "height"
	GeneBuild      = "build"
	GeneHair       = "hair"
	GenePigment    = "pigment"
	GeneEye        = "eye"
	GeneMelanin    = "melanin"
	GeneNose       = "nose"
	GeneJaw        = "jaw"
	GeneCheek      = "cheek"
)
