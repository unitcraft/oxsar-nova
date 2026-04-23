package economy

// Unit ID-константы для зданий и исследований.
// Источник: configs/construction.yml (поле id).
const (
	// Здания (mode=1)
	IDMetalmine        = 1
	IDSiliconLab       = 2
	IDHydrogenLab      = 3
	IDSolarPlant       = 4
	IDHydrogenPlant    = 5
	IDNanoFactory      = 7
	IDMetalStorage     = 9
	IDSiliconStorage   = 10
	IDHydrogenStorage  = 11
	IDRoboticFactory   = 6
	IDMoonHydrogenLab  = 326

	// Корабли/оборона
	IDSolarSatellite   = 39
	IDGravi            = 28

	// Исследования (mode=2): tech ID в formula.Context.Tech
	IDTechEnergy       = 18 // energy_tech   — energy_ratio
	IDTechLaser        = 23 // laser_tech     — metalmine prod
	IDTechSilicon      = 24 // silicon_tech   — silicon_lab prod
	IDTechHydrogen     = 25 // hydrogen_tech  — hydrogen_lab prod
)
