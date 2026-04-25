// Мини-каталог юнитов для UI. Значения дублируют configs/ships.yml + defense.yml.
// TODO: сгенерировать из YAML на этапе gen:api (см. CLAUDE.md).

export interface UnitEntry {
  id: number;
  key: string;
  name: string;
}

// CombatEntry — юнит с боевыми характеристиками для симулятора.
export interface CombatEntry extends UnitEntry {
  attack: number;
  shield: number;
  shell: number;
  cost?: Cost;
  cargo?: number;
  speed?: number;
  fuel?: number;
  requires?: Req[];
  description?: string;
  rapidfire?: Record<number, number>;
  // Из legacy ship_datasheet: боевые параметры для UnitInfoScreen
  front?: number;           // приоритет цели (выше = чаще атакуют первым)
  ballistics?: number;      // точность (уровень баллистики)
  masking?: number;         // маскировка (уровень)
  // attacker_* — значения когда юнит выступает в роли атакующего
  attacker_front?: number;
  attacker_ballistics?: number;
  attacker_masking?: number;
}

export interface Cost { metal: number; silicon: number; hydrogen: number }

export interface Req { kind: 'building' | 'research'; key: string; level: number }

export interface BuildingEntry extends UnitEntry {
  costBase: Cost;
  costFactor: number;
  description?: string;
  fullDesc?: string;
  maxLevel?: number;
}

// costForLevel: cost_base * cost_factor^(level-1), округление вниз.
export function costForLevel(base: Cost, factor: number, level: number): Cost {
  const m = factor ** (level - 1);
  return {
    metal:    Math.floor(base.metal    * m),
    silicon:  Math.floor(base.silicon  * m),
    hydrogen: Math.floor(base.hydrogen * m),
  };
}

// Форматирование чисел с сокращениями: 1.5M, 2k
export function formatNum(v: number): string {
  if (v >= 1_000_000) return (v / 1_000_000).toFixed(1) + 'M';
  if (v >= 1_000) return (v / 1_000).toFixed(0) + 'K';
  return Math.floor(v).toLocaleString('ru-RU');
}

export const BUILDINGS: BuildingEntry[] = [
  { id: 1,   key: 'metal_mine',       name: 'Рудник металла',                description: 'Добывает металл из недр планеты',                    fullDesc: 'Основной поставщик сырья для строительства несущих структур построек и кораблей. Металл — самое дешёвое сырьё, но зато его требуется больше, чем всего остального. Для производства металла требуется меньше всего энергии. Чем рудники больше, тем они глубже. На большинстве планет металл находится на больших глубинах, при помощи этих более глубоких рудников можно добывать больше металлов, производство растёт. В то же время более крупные рудники требуют больше энергии.',          costBase: { metal: 60,      silicon: 15,     hydrogen: 0      }, costFactor: 1.5 },
  { id: 2,   key: 'silicon_lab',      name: 'Рудник по добыче кремния',      description: 'Добывает кремний для строительства и исследований',   fullDesc: 'Основной поставщик сырья для электронных строительных элементов и сплавов. Для добычи кремния требуется примерно вдвое больше энергии, чем для добычи металла, поэтому он, соответственно, ценится больше. Кремний требуется для всех кораблей и зданий. К сожалению, так необходимые для строительства и исследований, месторождения кремния обычно очень редки и, как и большинство металлов, залегают на больших глубинах. Поэтому при усовершенствовании рудника также повышается производство, так как осваиваются более крупные и чистые залежи кремния.',      costBase: { metal: 48,      silicon: 24,     hydrogen: 0      }, costFactor: 1.6 },
  { id: 3,   key: 'hydrogen_lab',     name: 'Синтезатор водорода',            description: 'Синтезирует водород — топливо для флота',             fullDesc: 'Совершенствование синтезатора способствует увеличению его эффективности по обработке воды океана планеты и, как следствие, большей выработке водорода. Водород необходим в качестве топлива для кораблей, почти для всех исследований, просмотра галактик, а также для использования сенсорной фаланги.',             costBase: { metal: 225,     silicon: 75,     hydrogen: 0      }, costFactor: 1.5 },
  { id: 4,   key: 'solar_plant',      name: 'Солнечная электростанция',       description: 'Вырабатывает энергию из солнечного излучения',        fullDesc: 'Для обеспечения энергией рудников и синтезаторов необходимы огромные солнечные электростанции. Чем больше построено станций, тем больше поверхности покрыто кварцевыми пластинами, которые перерабатывают световую энергию в электроэнергию. Солнечные электростанции представляют собой основу энергообеспечения планеты.',        costBase: { metal: 75,      silicon: 30,     hydrogen: 0      }, costFactor: 1.5 },
  { id: 5,   key: 'hydrogen_plant',   name: 'Термоядерная электростанция',    description: 'Мощная электростанция на термоядерном синтезе',       fullDesc: 'На термоядерных электростанциях при помощи термоядерного синтеза под огромным давлением и при высокой температуре 2 атома тяжёлого водорода объединяются в один атом гелия. При этом вырабатывается огромное количество энергии. Чем больше термоядерный реактор, тем сложнее процессы синтезирования, реактор производит больше энергии.',       costBase: { metal: 900,     silicon: 360,    hydrogen: 180    }, costFactor: 1.8 },
  { id: 6,   key: 'robotic_factory',  name: 'Фабрика роботов',               description: 'Ускоряет строительство зданий',                       fullDesc: 'Предоставляет простую рабочую силу, которую можно применять при строительстве планетарной инфраструктуры. Каждый уровень развития фабрики повышает скорость строительства зданий.',                       costBase: { metal: 400,     silicon: 120,    hydrogen: 200    }, costFactor: 2.0 },
  { id: 7,   key: 'nano_factory',     name: 'Нано-фабрика',                  description: 'Вдвое ускоряет строительство за каждый уровень',      fullDesc: 'Фабрика нанитов представляет собой венец робототехники. Наниты — это роботы размером в нанометр, которые путём объединения в сеть в состоянии выполнять экстраординарные задания. С каждым уровнем фабрика нанитов сокращает время строительства зданий, кораблей и оборонительных сооружений вдвое.',      costBase: { metal: 1000000, silicon: 500000, hydrogen: 100000 }, costFactor: 2.0 },
  { id: 8,   key: 'shipyard',         name: 'Верфь',                         description: 'Производит корабли и оборонительные системы',         fullDesc: 'В строительной верфи производятся все виды кораблей. Чем она больше, тем быстрее можно строить более сложные и более крупные корабли. Посредством строительства фабрики нанитов производятся миниатюрные роботы, которые помогают работникам работать быстрее.',         costBase: { metal: 400,     silicon: 200,    hydrogen: 100    }, costFactor: 2.0 },
  { id: 9,   key: 'metal_storage',    name: 'Хранилище металла',             description: 'Увеличивает максимальный запас металла',              fullDesc: 'Огромное хранилище для добытых руд. Чем оно больше, тем больше металла можно в нём хранить. Если оно заполнено, то добыча металла прекращается.',              costBase: { metal: 1000,    silicon: 0,      hydrogen: 0      }, costFactor: 2.0 },
  { id: 10,  key: 'silicon_storage',  name: 'Хранилище кремния',             description: 'Увеличивает максимальный запас кремния',              fullDesc: 'В этом огромном хранилище складируется ещё не обработанный кремний. Чем больше хранилище, тем больше кремния туда помещается. Если оно заполнено, то добыча данного ресурса прекращается.',              costBase: { metal: 1000,    silicon: 500,    hydrogen: 0      }, costFactor: 2.0 },
  { id: 11,  key: 'hydrogen_storage', name: 'Емкость для водорода',           description: 'Увеличивает максимальный запас водорода',             fullDesc: 'Огромные ёмкости для хранения добытого водорода. Они обычно находятся вблизи космических портов. Чем они больше, тем больше водорода в них может сберегаться. Если они заполнены, то добыча водорода прекращается.',             costBase: { metal: 1000,    silicon: 1000,   hydrogen: 0      }, costFactor: 2.0 },
  { id: 12,  key: 'research_lab',     name: 'Исследовательская лаборатория', description: 'Позволяет проводить научные исследования',            fullDesc: 'Для исследования новых технологий необходима работа исследовательской станции. Уровень развития исследовательской станции является решающим фактором того, как быстро могут быть освоены новые технологии. Чем выше уровень развития исследовательской лаборатории, тем больше может быть исследовано новых технологий.',            costBase: { metal: 200,     silicon: 400,    hydrogen: 200    }, costFactor: 2.0 },
  { id: 53,  key: 'missile_silo',     name: 'Ракетная шахта',                description: 'Хранит межпланетные ракеты для атаки',                fullDesc: 'Ракетные шахты служат для хранения ракет. С каждым уровнем можно хранить на пять межпланетных или десять ракет-перехватчиков больше. Одна межпланетная ракета требует места в два раза больше, чем ракета-перехватчик. Возможно любое комбинирование различных типов ракет.',                costBase: { metal: 20000,   silicon: 20000,  hydrogen: 1000   }, costFactor: 2.0 },
  { id: 100, key: 'repair_factory',   name: 'Ремонтный ангар',               description: 'Восстанавливает повреждённые корабли после боя',      fullDesc: 'Ремонтный ангар необходим для выполнения двух операций: 1 — ремонт повреждённых кораблей и оборонительных сооружений, 2 — утилизация ненужных кораблей и обороны для получения ресурсов.',      costBase: { metal: 800,     silicon: 400,    hydrogen: 200    }, costFactor: 2.0 },
];

export const MOON_BUILDINGS: BuildingEntry[] = [
  { id: 54, key: 'moon_base',             name: 'Лунная база',             description: 'Основная постройка луны, даёт возможность строить другие здания',  fullDesc: 'Луна не располагает атмосферой, поэтому перед заселением требуется соорудить лунную базу. Она обеспечивает необходимые воздух, гравитацию и тепло. Чем выше уровень развития лунной базы, тем больше обеспеченная биосферой площадь. Каждый уровень лунной базы может застроить 3 поля, максимум до площади всей луны.',  costBase: { metal: 50000,    silicon: 20000,   hydrogen: 10000  }, costFactor: 2.0 },
  { id: 55, key: 'star_surveillance',     name: 'Звёздные сенсоры',       description: 'Следит за флотами противника в системе',                           costBase: { metal: 100000,   silicon: 20000,   hydrogen: 50000  }, costFactor: 2.0 },
  { id: 56, key: 'star_gate',             name: 'Звёздные врата',          description: 'Позволяет мгновенно переместить флот между лунами',                 fullDesc: 'Ворота — это огромные телепортеры, которые могут пересылать между собой флоты любых размеров без временных затрат.',                 costBase: { metal: 4000000,  silicon: 2000000, hydrogen: 1000000 }, costFactor: 2.0 },
  { id: 57, key: 'moon_robotic_factory',  name: 'Лунная фабрика роботов',  description: 'Ускоряет строительство зданий на луне',                           fullDesc: 'Предоставляет простую рабочую силу, которую можно применять при строительстве планетарной инфраструктуры. Каждый уровень развития фабрики повышает скорость строительства зданий.',                           costBase: { metal: 10000,    silicon: 6000,    hydrogen: 4000   }, costFactor: 2.0 },
];

export interface ResearchEntry extends UnitEntry {
  costBase: Cost;
  costFactor: number;
  benefit: string;
  fullDesc?: string;
  requires?: Req[];
}

export const RESEARCH: ResearchEntry[] = [
  { id: 13,  key: 'spyware',           name: 'Шпионаж',                              benefit: '+1 уровень шпионажа зонда',                fullDesc: 'Шпионаж предназначен для исследования новых и более эффективных сенсоров. Чем выше развита эта технология, тем больше информации имеет игрок о событиях в своём окружении. Разница в уровнях шпионажа с противником играет решающую роль — чем больше исследована собственная шпионская технология, тем больше информации содержится в разведданных и тем меньше шанс быть обнаруженным. Начиная со второго уровня при атаке на вас показывается также и общая численность нападающих кораблей. С четвёртого уровня распознаётся вид нападающих кораблей, а с восьмого — точная численность каждого типа кораблей.',                costBase: { metal: 200,   silicon: 1000,  hydrogen: 200  }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 3 }] },
  { id: 14,  key: 'computer_tech',     name: 'Компьютерная технология',              benefit: '+1 слот флота',                            fullDesc: 'Компьютерная технология предназначена для расширения имеющихся в наличии компьютерных мощностей. В результате на планете развиваются более продуктивные и эффективные компьютерные системы, возрастает вычислительная мощность и скорость протекания вычислительных процессов. С повышением мощности компьютеров можно одновременно командовать всё большим количеством флотов. Каждый уровень развития компьютерной технологии даёт возможность командовать +1 флотом.',                            costBase: { metal: 0,     silicon: 400,   hydrogen: 600  }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 1 }] },
  { id: 15,  key: 'gun_tech',          name: 'Оружейная технология',                 benefit: '+2% атака флота и обороны',                fullDesc: 'Оружейная технология занимается прежде всего дальнейшим развитием имеющихся в наличии систем вооружения. При этом особое значение придаётся тому, чтобы снабжать имеющиеся в наличии системы большей энергией и более точно эту энергию направлять. Благодаря этому системы вооружения становятся эффективней, а оружие вызывает больше разрушений. Каждый уровень оружейной технологии увеличивает мощность вооружения войсковых частей на 10%.',                costBase: { metal: 800,   silicon: 200,   hydrogen: 0    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 4 }] },
  { id: 16,  key: 'shield_tech',       name: 'Щитовая технология',                   benefit: '+2% щит флота и обороны',                  fullDesc: 'Развитие этой технологии позволяет увеличивать снабжение энергией щитов и защитных экранов, что в свою очередь повышает их устойчивость и способность поглощать или отражать энергию атак противника. Благодаря этому с каждым изученным уровнем эффективность корабельных щитов и стационарных генераторов энергополей повышается на 10% от номинальной мощности. Кроме этого, с каждым уровнем можно строить в обороне больше защитных куполов.',                  costBase: { metal: 200,   silicon: 600,   hydrogen: 0    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 6 }] },
  { id: 17,  key: 'shell_tech',        name: 'Броневая технология',                  benefit: '+2% броня флота и обороны',                fullDesc: 'Специальные сплавы улучшают броню космических кораблей. Как только найден очень стойкий сплав, специальные лучи изменяют молекулярную структуру космического корабля и доводят её до состояния изученного сплава. Так, устойчивость брони может увеличиваться с каждым уровнем на 10%.',                costBase: { metal: 1000,  silicon: 0,     hydrogen: 0    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 2 }] },
  { id: 18,  key: 'energy_tech',       name: 'Энергетическая технология',            benefit: 'требование для высоких технологий',        fullDesc: 'Обладание различными видами энергии необходимо для многих новых технологий.',        costBase: { metal: 0,     silicon: 800,   hydrogen: 400  }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 3 }] },
  { id: 19,  key: 'hyperspace_tech',   name: 'Гиперпространственная технология',     benefit: 'требование для гипердвигателя',            fullDesc: 'Путём сплетения 4-го и 5-го измерения стало возможным исследовать новый более экономный и эффективный двигатель. Кроме этого учёные обнаружили, что искривлённое пространство увеличивает радиус действия сенсорной фаланги на 10% за каждый уровень технологии.',            costBase: { metal: 0,     silicon: 4000,  hydrogen: 2000 }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 7 }, { kind: 'research', key: 'energy_tech', level: 5 }, { kind: 'research', key: 'shield_tech', level: 5 }] },
  { id: 20,  key: 'combustion_engine', name: 'Реактивный двигатель',                 benefit: '+10% скорость транспортов и истребителей', fullDesc: 'Реактивный двигатель основывается на принципе отдачи. Материя, разогретая до высоких температур, выбрасывается в направлении, противоположном движению и даёт ускорение кораблю. Эффективность этих двигателей достаточно мала, но они достаточно надёжны, дёшевы в производстве и обслуживании. Дальнейшее развитие этих двигателей делает малые транспорты, большие транспорты, лёгкие истребители и шпионские зонды с каждым уровнем на 10% быстрее.', costBase: { metal: 400,   silicon: 0,     hydrogen: 600  }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 1 }, { kind: 'research', key: 'energy_tech', level: 1 }] },
  { id: 21,  key: 'impulse_engine',    name: 'Импульсный двигатель',                 benefit: '+20% скорость крейсеров и зондов',         fullDesc: 'Импульсный двигатель основывается на принципе отдачи, причём разогрев материи осуществляется в ходе ядерной реакции. Дальнейшее развитие этих двигателей делает следующие корабли с каждым уровнем на 20% быстрее: бомбардировщики, крейсеры, тяжёлые истребители и колонизаторы. Каждый уровень развития увеличивает радиус действия межпланетных ракет.',         costBase: { metal: 2000,  silicon: 4000,  hydrogen: 600  }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 2 }, { kind: 'research', key: 'energy_tech', level: 1 }] },
  { id: 22,  key: 'hyperspace_engine', name: 'Гиперпространственный двигатель',      benefit: '+30% скорость линкоров и флагманов',       fullDesc: 'Благодаря пространственно-временному изгибу в непосредственном окружении корабля пространство сжимается, чем быстрее преодолеваются далёкие расстояния. Чем выше развит гиперпространственный привод, тем выше сжатие пространства, благодаря чему с каждым уровнем скорость кораблей повышается на 30%.',       costBase: { metal: 10000, silicon: 20000, hydrogen: 6000 }, costFactor: 3.0,  requires: [{ kind: 'building', key: 'research_lab', level: 7 }, { kind: 'research', key: 'hyperspace_tech', level: 3 }] },
  { id: 23,  key: 'laser_tech',        name: 'Лазерная технология',                  benefit: 'требование для ионной технологии',         fullDesc: 'Лазеры (усиление света при помощи индуцированного выброса излучения) производят насыщенный энергетический луч когерентного света. Эти приборы находят применение во всевозможных областях, от оптических компьютеров до тяжёлых лазеров, которые свободно режут броню космических кораблей. Лазерная технология является важным элементом для исследования дальнейших оружейных технологий.',         costBase: { metal: 200,   silicon: 100,   hydrogen: 0    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 1 }, { kind: 'research', key: 'energy_tech', level: 2 }] },
  { id: 24,  key: 'ion_tech',          name: 'Ионная технология',                    benefit: 'требование для плазменной технологии',     fullDesc: 'Поистине смертоносный наводимый луч из ускоренных ионов. При попадании на какой-либо объект они наносят огромный ущерб.',     costBase: { metal: 1000,  silicon: 300,   hydrogen: 100  }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 4 }, { kind: 'research', key: 'laser_tech', level: 5 }, { kind: 'research', key: 'energy_tech', level: 4 }] },
  { id: 25,  key: 'plasma_tech',       name: 'Плазменная технология',                benefit: 'повышенный урон по ресурсам противника',   fullDesc: 'Дальнейшее развитие ионной технологии, которая ускоряет не ионы, а высокоэнергетическую плазму. Она оказывает опустошительное действие при попадании на какой-либо объект.',   costBase: { metal: 2000,  silicon: 4000,  hydrogen: 1000 }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 4 }, { kind: 'research', key: 'ion_tech', level: 5 }, { kind: 'research', key: 'laser_tech', level: 8 }, { kind: 'research', key: 'energy_tech', level: 8 }] },
  { id: 27,  key: 'expo_tech',         name: 'Экспедиционная технология',            benefit: '+1 слот экспедиции за уровень',            fullDesc: 'Экспедиционная технология охватывает различные технологии сканирования и даёт возможность оснащать корабли различных классов исследовательским модулем. Он содержит базу данных, маленькую передвижную лабораторию, а также различные биоклетки и сосуды для проб. Для безопасности корабля при исследовании опасных объектов исследовательский модуль оснащён автономным энергообеспечением и генератором энергетического поля.',            costBase: { metal: 4000,  silicon: 8000,  hydrogen: 4000 }, costFactor: 1.75, requires: [{ kind: 'building', key: 'research_lab', level: 3 }, { kind: 'research', key: 'impulse_engine', level: 3 }, { kind: 'research', key: 'spyware', level: 4 }] },
  { id: 103, key: 'ballistics_tech',   name: 'Баллистическая технология',            benefit: '+1 ракета в шахте за уровень',             fullDesc: 'Технология баллистического анализа позволяет компьютерным системам просчитывать бой до его начала на основе пространственного расположения, технических характеристик юнитов и траекторий огневых путей и излучений. Это делает возможным вести прицельный и высокоточный огонь на всём протяжении боя. Технология уменьшает промахи и попадания в уже уничтоженные юниты.',             costBase: { metal: 4000,  silicon: 8000,  hydrogen: 4000 }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 5 }, { kind: 'research', key: 'gun_tech', level: 5 }] },
  { id: 104, key: 'masking_tech',      name: 'Маскировочная технология',             benefit: 'снижение видимости флота для шпионажа',    fullDesc: 'Технология радио-локационной маскировки создаёт помехи в работе баллистического анализатора и систем наведения огня противника. Эта технология увеличивает живучесть своих юнитов за счёт уменьшения точности стрельбы огневых орудий противника.',    costBase: { metal: 4000,  silicon: 8000,  hydrogen: 4000 }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 5 }, { kind: 'research', key: 'hyperspace_tech', level: 3 }] },
];

export const SHIPS: CombatEntry[] = [
  { id: 29, key: 'small_transporter', name: 'Малый транспорт',    attack: 5,      shield: 10,    shell: 4000,    cargo: 5000,    speed: 5000,      fuel: 10,   cost: { metal: 2000,    silicon: 2000,   hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 2 }, { kind: 'research', key: 'combustion_engine', level: 2 }], description: 'Дешёвый транспорт для перевозки ресурсов',                                          front: 10, ballistics: 0, masking: 0 },
  { id: 30, key: 'large_transporter', name: 'Большой транспорт',  attack: 5,      shield: 25,    shell: 12000,   cargo: 25000,   speed: 7500,      fuel: 50,   cost: { metal: 6000,    silicon: 6000,   hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 4 }, { kind: 'research', key: 'combustion_engine', level: 6 }], description: 'Основной грузовоз для перевозки ресурсов',                                             front: 10, ballistics: 0, masking: 0 },
  { id: 31, key: 'light_fighter',     name: 'Легкий истребитель', attack: 50,     shield: 10,    shell: 4000,    cargo: 50,      speed: 12500,     fuel: 20,   cost: { metal: 3000,    silicon: 1000,   hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 1 }, { kind: 'research', key: 'combustion_engine', level: 1 }], description: 'Дешёвый и быстрый — основа атакующего флота',                                        front: 10, ballistics: 0, masking: 0 },
  { id: 32, key: 'strong_fighter',    name: 'Тяжелый истребитель',attack: 150,    shield: 25,    shell: 10000,   cargo: 100,     speed: 10000,     fuel: 75,   cost: { metal: 6000,    silicon: 4000,   hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 3 }, { kind: 'research', key: 'impulse_engine', level: 2 }, { kind: 'research', key: 'shell_tech', level: 2 }], description: 'Мощная альтернатива лёгкому истребителю',                                            front: 10, ballistics: 0, masking: 0 },
  { id: 33, key: 'cruiser',           name: 'Крейсер',            attack: 400,    shield: 50,    shell: 27000,   cargo: 800,     speed: 15000,     fuel: 300,  cost: { metal: 20000,   silicon: 7000,   hydrogen: 2000    }, requires: [{ kind: 'building', key: 'shipyard', level: 5 }, { kind: 'research', key: 'impulse_engine', level: 4 }, { kind: 'research', key: 'ion_tech', level: 2 }], description: 'Эффективен против ракетных установок',                                               front: 10, ballistics: 0, masking: 0, rapidfire: { 31: 6, 43: 10 } },
  { id: 34, key: 'battle_ship',       name: 'Линкор',             attack: 1000,   shield: 200,   shell: 60000,   cargo: 1500,    speed: 10000,     fuel: 500,  cost: { metal: 45000,   silicon: 15000,  hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 7 }, { kind: 'research', key: 'hyperspace_engine', level: 4 }], description: 'Мощный боевой корабль с гиперпространственным двигателем',                           front: 10, ballistics: 0, masking: 0, rapidfire: { 38: 5, 39: 5 } },
  { id: 36, key: 'colony_ship',       name: 'Колонизатор',        attack: 50,     shield: 100,   shell: 30000,   cargo: 7500,    speed: 2500,      fuel: 1000, cost: { metal: 10000,   silicon: 20000,  hydrogen: 10000   }, requires: [{ kind: 'building', key: 'shipyard', level: 4 }, { kind: 'research', key: 'impulse_engine', level: 3 }], description: 'Позволяет колонизировать незанятые планеты',                                         front: 10, ballistics: 0, masking: 0 },
  { id: 37, key: 'recycler',          name: 'Переработчик',       attack: 1,      shield: 10,    shell: 16000,   cargo: 20000,   speed: 2000,      fuel: 300,  cost: { metal: 12500,   silicon: 2500,   hydrogen: 10000   }, requires: [{ kind: 'building', key: 'shipyard', level: 4 }, { kind: 'research', key: 'combustion_engine', level: 6 }, { kind: 'research', key: 'shield_tech', level: 2 }], description: 'Собирает ресурсы из полей обломков',                                                 front: 10, ballistics: 0, masking: 0 },
  { id: 38, key: 'espionage_sensor',  name: 'Шпионский зонд',     attack: 0,      shield: 0,     shell: 1000,    cargo: 5,       speed: 100000000, fuel: 1,    cost: { metal: 0,       silicon: 1000,   hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 3 }, { kind: 'research', key: 'combustion_engine', level: 3 }, { kind: 'research', key: 'spyware', level: 2 }], description: 'Разведывает планеты — при слабом шпионаже может быть перехвачен',                    front: 10, ballistics: 0, masking: 0 },
  { id: 39, key: 'solar_satellite',   name: 'Солнечный спутник',  attack: 1,      shield: 1,     shell: 2000,                    speed: 5000,      fuel: 0,    cost: { metal: 0,       silicon: 2000,   hydrogen: 500     }, requires: [{ kind: 'building', key: 'shipyard', level: 1 }], description: 'Добавляет энергию без строительства электростанций',                                 front: 10, ballistics: 0, masking: 0 },
  { id: 40, key: 'bomber',            name: 'Бомбардировщик',     attack: 1000,   shield: 500,   shell: 75000,   cargo: 500,     speed: 4000,      fuel: 700,  cost: { metal: 50000,   silicon: 25000,  hydrogen: 15000   }, requires: [{ kind: 'building', key: 'shipyard', level: 8 }, { kind: 'research', key: 'impulse_engine', level: 6 }, { kind: 'research', key: 'plasma_tech', level: 5 }], description: 'Специализируется на уничтожении оборонительных сооружений',                          front: 10, ballistics: 0, masking: 0 },
  { id: 42, key: 'death_star',        name: 'Звезда смерти',      attack: 200000, shield: 50000, shell: 9000000, cargo: 1000000, speed: 100,       fuel: 1,    cost: { metal: 5000000, silicon: 4000000,hydrogen: 1000000 }, requires: [{ kind: 'building', key: 'shipyard', level: 12 }, { kind: 'research', key: 'hyperspace_tech', level: 6 }, { kind: 'research', key: 'hyperspace_engine', level: 7 }], description: 'Сильнейший корабль — способен уничтожить луну',                                      front: 10, ballistics: 4, masking: 0, attacker_front: 9, attacker_ballistics: 4, attacker_masking: 0, rapidfire: { 29: 250, 30: 250, 31: 200, 32: 100, 33: 33, 34: 30, 37: 250, 38: 1250, 39: 1250, 40: 25, 41: 5, 43: 200, 44: 200, 45: 100, 46: 100, 47: 50, 48: 50 } },
];

export const DEFENSE: CombatEntry[] = [
  { id: 43, key: 'rocket_launcher', name: 'Ракетная установка',   attack: 80,   shield: 20,    shell: 2000,    cost: { metal: 2000,  silicon: 0,    hydrogen: 0    }, requires: [{ kind: 'building', key: 'shipyard', level: 1 }], description: 'Базовая и дешёвая оборонительная установка',  front: 10, ballistics: 0, masking: 0 },
  { id: 44, key: 'light_laser',     name: 'Легкий лазер',         attack: 100,  shield: 25,    shell: 2000,    cost: { metal: 1500,  silicon: 500,  hydrogen: 0    }, requires: [{ kind: 'building', key: 'shipyard', level: 2 }, { kind: 'research', key: 'laser_tech', level: 3 }], description: 'Лазерная пушка начального уровня',            front: 10, ballistics: 0, masking: 0 },
  { id: 45, key: 'strong_laser',    name: 'Тяжелый лазер',        attack: 250,  shield: 100,   shell: 8000,    cost: { metal: 6000,  silicon: 2000, hydrogen: 0    }, requires: [{ kind: 'building', key: 'shipyard', level: 4 }, { kind: 'research', key: 'laser_tech', level: 6 }, { kind: 'research', key: 'energy_tech', level: 3 }], description: 'Усиленная лазерная пушка с большей мощностью', front: 10, ballistics: 0, masking: 1 },
  { id: 47, key: 'gauss_gun',       name: 'Пушка Гаусса',         attack: 1100, shield: 200,   shell: 35000,   cost: { metal: 20000, silicon: 15000,hydrogen: 2000 }, requires: [{ kind: 'building', key: 'shipyard', level: 6 }, { kind: 'research', key: 'gun_tech', level: 3 }, { kind: 'research', key: 'shield_tech', level: 1 }, { kind: 'research', key: 'energy_tech', level: 6 }], description: 'Мощная пушка — эффективна против тяжёлых кораблей', front: 10, ballistics: 1, masking: 2 },
  { id: 48, key: 'plasma_gun',      name: 'Плазменное орудие',    attack: 3000, shield: 300,   shell: 100000,  cost: { metal: 50000, silicon: 50000,hydrogen: 30000}, requires: [{ kind: 'building', key: 'shipyard', level: 8 }, { kind: 'research', key: 'plasma_tech', level: 7 }], description: 'Наиболее разрушительное орудие обороны',       front: 10, ballistics: 2, masking: 2 },
  { id: 49, key: 'small_shield',    name: 'Малый щитовой купол',  attack: 1,    shield: 2000,  shell: 20000,   cost: { metal: 10000, silicon: 10000,hydrogen: 0    }, requires: [{ kind: 'building', key: 'shipyard', level: 1 }, { kind: 'research', key: 'shield_tech', level: 2 }], description: 'Защищает всю оборону планеты от одного залпа',  front: 16, ballistics: 0, masking: 0 },
  { id: 50, key: 'large_shield',    name: 'Большой щитовой купол',attack: 1,    shield: 10000, shell: 100000,  cost: { metal: 50000, silicon: 50000,hydrogen: 0    }, requires: [{ kind: 'building', key: 'shipyard', level: 6 }, { kind: 'research', key: 'shield_tech', level: 6 }], description: 'Усиленный купол — щит в 5× больше малого',     front: 18, ballistics: 0, masking: 0 },
];

// Артефакты — только те, что реально реализованы в M5.0.1 (факторы).
// Остальные 300-365 добавятся в M5.1 вместе с one_shot/battle_bonus.
interface ArtefactEntry extends UnitEntry { benefit: string; lifetime: string }
export const ARTEFACTS: ArtefactEntry[] = [
  { id: 300, key: 'merchants_mark',       name: 'Знак торговца',                benefit: '+3% курс обмена ресурсов',          lifetime: '7 дней' },
  { id: 301, key: 'catalyst',             name: 'Катализатор',                  benefit: '+10% добыча на всех планетах',       lifetime: '7 дней' },
  { id: 302, key: 'power_generator',      name: 'Энерготранс',                  benefit: '+15% энергия на всех планетах',      lifetime: '7 дней' },
  { id: 303, key: 'atomic_densifier',     name: 'Атомный уплотнитель',          benefit: '+15% ёмкость склада на всех планетах',lifetime: '7 дней' },
  { id: 305, key: 'supercomputer',        name: 'Суперкомпьютер',               benefit: '+100% скорость исследования',        lifetime: '7 дней' },
  { id: 315, key: 'robot_control_system', name: 'Система управления роботами',  benefit: '+100% скорость строительства (планета)', lifetime: '7 дней' },
];

const KEY_MAP: Record<string, string> = {
  metal_mine:   'metalmine',
  missile_silo: 'rocket_station',
};

// imageOf возвращает путь к иконке юнита из /images/units/ (legacy std skin).
export function imageOf(key: string): string {
  return `/images/units/${KEY_MAP[key] ?? key}.gif`;
}

// imageOfId возвращает путь к иконке юнита по его числовому id.
export function imageOfId(id: number): string {
  for (const c of [SHIPS, DEFENSE, RESEARCH, ARTEFACTS, BUILDINGS]) {
    const u = c.find((x) => x.id === id);
    if (u) return imageOf(u.key);
  }
  return '';
}

// nameByKey возвращает отображаемое имя юнита по ключу.
function nameByKey(key: string): string {
  for (const c of [...BUILDINGS, ...RESEARCH, ...SHIPS, ...DEFENSE]) {
    if (c.key === key) return c.name;
  }
  return key;
}

// fmtReqs форматирует список требований в читаемую строку.
export function fmtReqs(reqs: Req[]): string {
  return reqs.map((r) => `${nameByKey(r.key)} ур.${r.level}`).join(' + ');
}

// Тип планеты по позиции в системе (из PlanetPictures.xml legacy).
const PLANET_TYPES: Array<{ name: string; from: number; to: number; count: number }> = [
  { name: 'trockenplanet',    from: 1,  to: 4,  count: 10 },
  { name: 'wuestenplanet',    from: 1,  to: 3,  count: 4  },
  { name: 'dschjungelplanet', from: 3,  to: 7,  count: 10 },
  { name: 'normaltempplanet', from: 6,  to: 10, count: 7  },
  { name: 'wasserplanet',     from: 9,  to: 13, count: 9  },
  { name: 'eisplanet',        from: 12, to: 15, count: 10 },
  { name: 'gasplanet',        from: 13, to: 15, count: 8  },
];

// planetImageOf возвращает путь к картинке планеты.
// Если передан planetType — используется напрямую (из БД).
// Иначе — детерминировано из позиции слота и хэша id планеты.
export function planetImageOf(position: number, planetId: string, planetType?: string): string {
  // простой хэш из id планеты
  let h = 0;
  for (let i = 0; i < planetId.length; i++) h = (h * 31 + planetId.charCodeAt(i)) >>> 0;

  let type: (typeof PLANET_TYPES)[number];
  if (planetType && planetType !== 'moon') {
    const found = PLANET_TYPES.find((t) => t.name === planetType);
    type = found ?? PLANET_TYPES[3]!;
  } else {
    const eligible = PLANET_TYPES.filter((t) => position >= t.from && position <= t.to);
    const types = eligible.length > 0 ? eligible : [PLANET_TYPES[3]!];
    type = types[h % types.length]!;
  }
  const num = (h % type.count) + 1;
  return `/images/planets/${type.name}${String(num).padStart(2, '0')}.jpg`;
}

// planetImageSize возвращает размер в пикселях для отображения планеты по диаметру.
// Диапазон диаметров: ~2000 (луна) до ~17000. Планеты: 32px..64px.
export function planetImageSize(diameter?: number): number {
  if (!diameter) return 48;
  return 32 + Math.round((Math.min(diameter, 17000) / 17000) * 32);
}

export function nameOf(id: number): string {
  for (const c of [SHIPS, DEFENSE, RESEARCH, ARTEFACTS, BUILDINGS]) {
    const u = c.find((x) => x.id === id);
    if (u) return u.name;
  }
  return `#${id}`;
}

// keyOfId — ключ юнита по числовому id (для wiki-навигации, slug страницы).
export function keyOfId(id: number): string | null {
  for (const c of [SHIPS, DEFENSE, RESEARCH, ARTEFACTS, BUILDINGS]) {
    const u = c.find((x) => x.id === id);
    if (u) return u.key;
  }
  return null;
}

// categoryOfId — wiki-категория для unit_id. ships/defense/buildings/research.
export function categoryOfId(id: number): string | null {
  if (SHIPS.find((x) => x.id === id)) return 'ships';
  if (DEFENSE.find((x) => x.id === id)) return 'defense';
  if (BUILDINGS.find((x) => x.id === id)) return 'buildings';
  if (RESEARCH.find((x) => x.id === id)) return 'research';
  return null;
}

export function buildingName(id: number): string {
  return BUILDINGS.find((b) => b.id === id)?.name ?? `#${id}`;
}

// API functions for resource management
import { api } from './client';
import type { ResourceReport } from './types';

export const resourceAPI = {
  getResourceReport: (planetId: string) =>
    api.get<ResourceReport>(`/api/planets/${planetId}/resource-report`),

  updateResourceFactors: (planetId: string, payload: { factors: Record<string, number> }) =>
    api.post<{ status: string }>(`/api/planets/${planetId}/resource-update`, payload),
};

// Re-export for convenience
export const catalog = {
  getResourceReport: resourceAPI.getResourceReport,
  updateResourceFactors: resourceAPI.updateResourceFactors,
};
