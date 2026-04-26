<?php
/**
* Initialize language system.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Language.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Language extends Collection
{
	/**
	* Language ID.
	*
	* @var mixed
	*/
	protected $langid;

	/**
	* Variable names.
	*
	* @var array
	*/
	protected $vars = array();

	/**
	* Phrases groups in usage.
	*
	* @var array
	*/
	protected $grouplist = array();

	/**
	* Particiluar options of current language.
	*
	* @var array
	*/
	protected $opts = array();

	/**
	* Short language code. Indentified with two letters.
	*
	* @var string
	*/
	protected $langcode = "";

	/**
	* Enable cache.
	*
	* @var boolean
	*/
	protected $cacheActive = false;

	/**
	* Enables automatic cache rebuilding.
	*
	* @var boolean
	*/
	protected $autoCaching = true;

	/**
	* Indicates if all variables were already cached.
	*
	* @var boolean
	*/
	protected $uncached = true;

	/**
	* Holds dynamic variables.
	*
	* @var array
	*/
	protected $dynvars = array();

	protected $all_groups = array();
	
	/**
	* Constructor.
	*
	* @param string	Shortcut of language package.
	*
	* @return void
	*/
	public function __construct($langid, $groups)
	{
		$this->langid = $langid;
		$this->getDefaultLang();
		$this->item = array();
		$this->vars = array();
		if( $this->cacheActive )
		{
			$this->cacheActive = CACHE_ACTIVE;
		}
		$this->loadAllGroups();
		// $this->opts = $this->getOptions();
		$this->load($groups);
		return;
	}
	
	protected function getLoadLangs()
	{
		$langs = array(DEF_LANGUAGE_ID);
		// if( ($lang_id = NS::getLanguageId()) != DEF_LANGUAGE_ID )
		if($this->langid != DEF_LANGUAGE_ID)
		{
			$langs[] = $this->langid; // $lang_id;
		}
		return $langs;
	}
	
	protected function loadAllGroups()
	{
		$c_name = 'Language.loadAllGroups'; // .NS::getLanguageId();
		$groups = false;
		if( $groups === false )
		{
			$groups = array();
			// foreach($this->getLoadLangs() as $lang_id)
			{
				$res = Core::getDB()->query(
					"SELECT phrasegroupid, title FROM `" . PREFIX . "phrasesgroups` ORDER BY phrasegroupid ASC"
				);
				while($row = Core::getDB()->fetch($res))
				{
					$groups[$row['title']]['id']     = $row['phrasegroupid'];
					$groups[$row['title']]['loaded'] = false;
				}
			}
			// cache disabled
		}
		$this->all_groups = $groups;
	}

	/**
	* Set phrases groups.
	*
	* @param string	List of groups, separated by comma
	*
	* @return array	All language variables
	*/
	public function load($groups = array())
	{
		if(!is_array($groups))
		{
			$groups = Arr::trimArray(explode(",", $groups));
		}
		return $this->loadGroups($groups);
//		$this->grouplist = array_merge($this->grouplist, $groups);
//		$this->grouplist = array_unique($this->grouplist);
//		sort($this->grouplist);
//		return $this->fetch();
	}
	
	/**
	 * Loads groups
	 * @param array $groups Group Name as value
	 */
	protected function loadGroups( $groups )
	{
		if( empty($groups) || !is_array($groups) )
		{
			return ;
		}
		foreach( $groups as $group )
		{
			if(
				isset($this->all_groups[$group]) && $this->all_groups[$group]['loaded'] == false
					&& isset($this->all_groups[$group]['id']) && !empty($this->all_groups[$group]['id'])
			)
			{
				$this->loadGroup($this->all_groups[$group]['id']);
				$this->all_groups[$group]['loaded'] = true;
			}
		}
	}
	
	/**
	 * Loads All phrases of single group.
	 * @param unknown_type $id
	 */
	protected function loadGroup($id)
	{
		$load_langs = $this->getLoadLangs();
		$c_name = 'Language.loadGroup:'.$id.','.end($load_langs);
		$phrases =false;
		if( $phrases === false )
		{
			$phrases = array();
			foreach($load_langs as $lang_id)
			{
				$res = Core::getDB()->query(
					"SELECT title, content FROM `" . PREFIX . "phrases`"
					. " WHERE phrasegroupid = " . (int)$id
					. " AND languageid = " . (int)$lang_id
				);
				while($row = Core::getDB()->fetch($res))
				{
					$compiler = new LanguageCompiler($row['content'], true);
					$phrases[$row['title']] = $compiler->getPhrase();
					$compiler->shutdown();
				}
			}
			// cache disabled
		}
		foreach( $phrases as $title => $content )
		{
			$this->setItem($title, $content);
		}
	}

	/**
	* Load language variables into an array.
	*
	* @return Language
	*/
	protected function fetch()
	{
		if($this->cacheActive)
		{
			$this->item = array_merge($this->item, Core::getCache()->getLanguageCache($this->opts["langcode"], $this->grouplist));
			$this->vars = array();
			foreach($this->item as $key => $val)
			{
				$this->item[$key] = $this->parse($val);
				array_push($this->vars, $key);
			}
		}
		else
		{
			$this->getFromDB();
		}
		return $this;
	}

	protected $pg_pulled = array();
	
	/**
	* Load language variables into array.
	*
	* @return Language
	*/
	protected function getFromDB()
	{
		$whereclause = array();
		for($i = 0; $i < count($this->grouplist); $i++)
		{
			if( !isset($this->pg_pulled[$this->grouplist[$i]]) )
			{
				$whereclause[] = "pg.title = ".sqlVal($this->grouplist[$i]);
				$this->pg_pulled[$this->grouplist[$i]] = true;
				// if($i < count($this->grouplist) - 1) { $whereclause .= " OR "; }
			}
		}
		if( empty($whereclause) )
		{
			return $this;
		}
		$whereclause = "p.languageid = ".sqlVal($this->langid)." AND (" . implode(" OR ", $whereclause).")";
		$select = array("p.title", "p.content");
		$result = Core::getQuery()->select("phrases AS p", $select, "LEFT JOIN ".PREFIX."phrasesgroups AS pg ON (pg.phrasegroupid = p.phrasegroupid)", $whereclause);
		while($row = Core::getDatabase()->fetch($result))
		{
			$compiler = new LanguageCompiler($row["content"], true);
			$row["content"] = $compiler->getPhrase();
			$compiler->shutdown();
			$this->setItem($row["title"], $row["content"]);
		}
		Core::getDatabase()->free_result($result);
		return $this;
	}

	/**
	* Returns all var names.
	*
	* @return array	Var names
	*/
	public function getVarnames()
	{
		return $this->vars;
	}

	/**
	* Parses a language variable for global variables.
	*
	* @param string	The variable content
	*
	* @return string	The parsed variable content
	*/
	protected function parse($langvar)
	{
		return addcslashes($langvar, '"\\');
	}

	/**
	* Counts array size.
	*
	* @return integer	The number of array elements
	*/
	public function getVarCount()
	{
		return sizeof($this->item);
	}

	/**
	* Sets all language options.
	*
	* @return array	Options
	*/
	public function getOptions()
	{
//		try {
			$row = Core::getDatabase()->query_unique("SELECT * FROM ".PREFIX."languages WHERE languageid = '".$this->langid."'");
//		} catch(Exception $e) { $e->printError(); }
		return $row;
	}

	/**
	* Detects default language.
	*
	* @return Language
	*/
	public function getDefaultLang()
	{
		if(Str::length($this->langid) == 0)
		{
			$this->langcode = $_SERVER["HTTP_ACCEPT_LANGUAGE"][0].$_SERVER["HTTP_ACCEPT_LANGUAGE"][1];
			$this->langid = 0;
		}
		else if(!is_numeric($this->langid))
		{
			$this->langcode = $this->langid;
			$this->langid = 0;
		}
		$this->opts = array();
		if($this->langid)
		{
			$this->opts = sqlSelectRow("languages", "*", "", "languageid = ".sqlVal($this->langid));
		}
		if(!$this->opts)
		{
			$this->opts = sqlSelectRow("languages", "*", "", "langcode = ".sqlVal($this->langcode));
		}
		if(!$this->opts)
		{
			$this->opts = sqlSelectRow("languages", "*", "", "languageid = ".sqlVal(Core::getOptions()->defaultlanguage));
		}
		$this->langid = $this->opts["languageid"];
		return $this;
	}

	/**
	* Returns item. If item does not exist, return item name.
	*
	* @param string	Variable name
	*
	* @return string	Value
	*/
	public function getItem($var, & $tpl = null)
	{
		// Hook::event("GET_LANGUAGE_ITEM", array(&$this, &$var));
		if($this->exists($var))
		{
			return $this->replaceWildCards($this->item[$var], $tpl);
		}
		// Try to cache language.
		if(
			$this->cacheActive &&
			$this->autoCaching &&
			$this->uncached
		)
		{
			Core::getCache()->cacheLanguage($this->opts["langcode"]);
		}
		$this->uncached = false;
		return $var;
	}

	public function getItemWith($var, $params)
	{
		return $this->getItem($var, $params);
	}

	/**
	* Alias to getItem().
	*
	* @param string	Variable name
	*
	* @return string	Value
	*/
	public function get($var, & $tpl = null)
	{
		return $this->getItem($var, $tpl);
	}

	/**
	* Set and fill item with content.
	*
	* @param string	Var name
	* @param string	Value
	*
	* @return Language
	*/
	public function setItem($var, $value)
	{
		if(!$this->exists($var))
		{
			array_push($this->vars, $var);
		}
		$this->item[$var] = $this->parse($value);
		return $this;
	}

	/**
	* Alias to setItem().
	*
	* @param string	Var name
	* @param string	Value
	*
	* @return Language
	*/
	public function set($var, $value)
	{
		return $this->setItem($var, $value);
	}

	/**
	* Assigns values to phrases variables.
	*
	* @param mixed		The variable name
	* @param mixed		The value to assign
	*
	* @return Language
	*/
	public function assign($variable, $value = null)
	{
		if(is_array($variable))
		{
			foreach($variable as $key => $val)
			{
				if(Str::length($key) > 0) { $this->assign($key, $val); }
			}
		}
		else if(is_string($variable) || is_numeric($variable))
		{
			if(Str::length($variable) > 0) { $this->dynvars[$variable] = $value; }
		}
		return $this;
	}

	/**
	* Replaces wildcards like {@assignment}.
	*
	* @param string	String to parse
	*
	* @return string	parsed string
	*/
	protected function replaceWildCards($content, & $tpl = null)
	{
		$_dynvars = &$this->dynvars;
		return preg_replace_callback("/\{\@([^\"]+)}/siU", function($m) use($tpl, &$_dynvars) {
			$key = $m[1];
			if(is_array($tpl) && isset($tpl[$key])) return $tpl[$key];
			if(isset($_dynvars[$key])) return $_dynvars[$key];
			if(is_object($tpl)) return $tpl->get($key);
			return '';
		}, $content);
	}

	/**
	* Rebuilds the language cache for the current language.
	*
	* @param mixed		Indicates a special phrase group
	*
	* @return Language
	*/
	public function rebuild($groups = null)
	{
		if(!is_null($groups))
		{
			Core::getCache()->cachePhraseGroup($groups, $this->opts["langcode"]);
			return;
		}
		Core::getCache()->cacheLanguage($this->opts["langcode"]);
		return $this;
	}

	/**
	* Returns an option parameter for the language.
	*
	* @param string	Parameter name
	*
	* @return mixed	Parameter value
	*/
	public function getOpt($param)
	{
		return (isset($this->opts[$param])) ? $this->opts[$param] : $param;
	}
}
?>