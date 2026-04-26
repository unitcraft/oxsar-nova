<?php
/**
* On Smarty based template engine.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Template.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR."game/Functions.inc.php");

class Template
{
  /**
  * File Extension.
  *
  * @var string
  */
  protected $templateExtension = "";

  /**
  * Name of main template file.
  *
  * @var string
  */
  protected $mainTemplateFile = "";

  /**
  * Absolute path to template directory.
  *
  * @var string
  */
  protected $templatePath = "";
  protected $templatePath2 = "";

  /**
  * Absolute path to template cache directory.
  *
  * @var string
  */
  protected $templateCachePath = "";

  /**
  * Package directory.
  *
  * @var string
  */
  protected $templatePackage = "";

  /**
  * Holds template variables.
  *
  * @var array
  */
  protected $templateVars = array();

  /**
  * If compilation will always performed.
  *
  * @var boolean
  */
  protected $forceCompilation = false;

  /**
  * Loop stack.
  *
  * @var array
  */
  protected $loopStack = array();

  // protected $runLoopStack = array();
  // protected $runKeyStack = array();
  protected $runValueStack = array();
  protected $runTempStack = array();

  /**
  * Contains log messages.
  *
  * @var string
  */
  protected $log = null;

  /**
  * If out stream is compressed.
  *
  * @var boolean
  */
  protected $compressed = false;

  /**
  * If HTTP header has been sent.
  *
  * @var boolean
  */
  protected $headersSent = false;

  /**
  * Holds html head data.
  *
  * @var Map
  */
  protected $htmlHead = null;
  protected $headKeys = array();

  /**
  * The view object.
  *
  * @var object
  */
  protected $view = null;

  /**
  * Constructor. Set default values.
  *
  * @return void
  */
  function __construct()
  {
    $this->setTemplatePath(APP_ROOT_DIR."ext/templates/");
    $this->setTemplatePath2(APP_ROOT_DIR."templates/");
    $this->setTemplatePackage();
    $this->setLayoutTemplate();
    $this->setExtension();
    $this->forceCompilation = false;
    $this->log = new Map();
    $this->htmlHead = new Map();
    $this->assign("LOG", "");
    $this->assign("HTML_HEAD", "");
    return;
  }

  /**
  * Assigns values to template variables.
  *
  * @param mixed		The template variable names
  * @param mixed		The value to assign
  *
  * @return Template
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
      if(Str::length($variable) > 0) { $this->templateVars[$variable] = $value; }
    }
    return $this;
  }

  /**
  * Clears the given assigned template variable.
  *
  * @param string	The template variable to flush
  *
  * @return Template
  */
  public function deallocateAssignment($variable)
  {
    if(is_array($variable))
    {
      foreach($variable as $v)
      {
        $this->deallocateAssignment($v);
      }
    }
    else
    {
      unset($this->templateVars[$variable]);
    }
    return $this;
  }

  /**
  * Clears assignment of all template variables.
  *
  * @return Template
  */
  public function deallocateAllAssignment()
  {
    $this->templateVars = array();
    return $this;
  }

    public function processOutput($output)
    {
        if(true){
            $output = preg_replace("#[ \t]+#", " ", $output);
            $output = preg_replace("#[\r\n]+ +[\r\n]+|[\r\n]+ +| +[\r\n]+#", "\n", $output);
            $output = preg_replace("#>\s+<#", "> <", $output);
        }
        echo $output;
    }

  /**
  * Executes and displays the template results.
  *
  * @param string	The to displayed template
  * @param boolean	For AJAX requests
  *
  * @return Template
  */
  function display($template, $sendOnlyContent = false, $mainTemplate = null, $use_once = true)
  {
  	if( defined('LOCAL') )
  	{
  		t($template);
  	}
    //user styles
    if(!$sendOnlyContent && !Core::getUser()->isGuest())
    {
      foreach(array("bg", "table") as $type)
      {
        $userStyle = Core::getUser()->get("user_{$type}_style");
        $styles = getUserStyles($type);
        if (isset($styles[$userStyle]))
          $this->addHTMLHeaderFile($styles[$userStyle]["path"]."?".CLIENT_VERSION, "css");
      }
    }

    $view = $this->getView();
    if($this->log->size() > 0)
    {
      $this->assign("LOG", $this->log->toString("\n"));
    }
    if($this->htmlHead->size() > 0)
    {
      $this->assign("HTML_HEAD", $this->htmlHead->toString("\n"));
    }
    $this->sendHeader();
    // Hook::event("TEMPLATE_START_OUTSTREAM", array(&$this, &$template, $sendOnlyContent, &$mainTemplate));
    if(!$sendOnlyContent || $mainTemplate != null)
    {
        if($mainTemplate == null) { $mainTemplate = $this->mainTemplateFile; }
        if(!$this->cachedTemplateAvailable($mainTemplate) || $this->forceCompilation)
        {
            new TemplateCompiler($this->getTemplatePath($mainTemplate));
        }
        ob_start();
        require_once(Core::getCache()->getTemplatePath($mainTemplate));
        $this->processOutput(ob_get_clean());
    }
    if(!$this->cachedTemplateAvailable($template) || $this->forceCompilation)
    {
      new TemplateCompiler($this->getTemplatePath($template));
    }
    $filename = Core::getCache()->getTemplatePath($template);

    ob_start();
    if($use_once)
    {
      require_once($filename);
    }
    else
    {
      require($filename);
    }
    $this->processOutput(ob_get_clean());
    return $this;
  }

  /**
  * Checks if template needs to be compiled.
  *
  * @param string	Template name
  *
  * @return boolean
  */
  protected function cachedTemplateAvailable($template)
  {
    $cached = Core::getCache()->getTemplatePath($template);
    $template = $this->getTemplatePath($template);
    if(!file_exists($template))
    {
      throw new GenericException($template." does not exist.");
      return false;
    }
    if(!$this->forceCompilation && file_exists($cached))
    {
      if(filemtime($template) <= filemtime($cached))
      {
        return true;
      }
    }
    return false;
  }

  /**
  * Return full path of a template.
  *
  * @param string	Template name
  *
  * @return string	Path to template
  */
  protected function getTemplatePath($template)
  {
    $template_pkg = $this->getTemplatePackage();
    // check main path
    $filename = $this->templatePath.$template_pkg.$template.$this->templateExtension;
    if(file_exists($filename))
    {
      return $filename;
    }
    if($this->templatePath2)
    {
      // check generic path
      $filename = $this->templatePath2.$template_pkg.$template.$this->templateExtension;
      if(file_exists($filename))
      {
        return $filename;
      }
      // check main default path
      $filename = $this->templatePath."standard/".$template.$this->templateExtension;
      if(file_exists($filename))
      {
        return $filename;
      }
      // return generic default path
      return $this->templatePath2."standard/".$template.$this->templateExtension;
    }
    return $this->templatePath."standard/".$template.$this->templateExtension;
  }

  /**
  * Append a loop to loop stack.
  *
  * @param array		Loop data
  *
  * @return Template
  */
  public function addLoop($loop, $data)
  {
    $this->loopStack[$loop] = $data;
    /* foreach($data as $key => $value)
    {
      if(is_array($value))
      {
        $this->addLoop("{$loop}.{$key}", $value);
      }
    } */
    return $this;
  }

  /**
  * Includes a template.
  *
  * @param string	Template name
  *
  * @return Template
  */
  public function includeTemplate($template)
  {
    $this->display($template, true, null, false);
    return $this;
  }

  /**
  * Sends header information to client and compress output.
  *
  * @return Template
  */
  protected function sendHeader()
  {
    if(!headers_sent() || !$this->headersSent)
    {
      if(@extension_loaded('zlib') && !$this->compressed && GZIP_ACITVATED)
      {
        ob_start("ob_gzhandler");
        $this->compressed = true;
      }
      @header("Content-Type: text/html; charset=".Core::getLanguage()->getOpt("charset"));
      $this->headersSent = true;
    }
    return $this;
  }

  /**
  * Sets a new template path.
  *
  * @param string	Path
  *
  * @return Template
  */
  public function setTemplatePath($path)
  {
    if(Str::substring($path, 0, Str::length($path) - 1) != "/")
    {
      $path .= "/";
    }
    $this->templatePath = $path;
    return $this;
  }

  /**
  * Sets a new template path.
  *
  * @param string	Path
  *
  * @return Template
  */
  public function setTemplatePath2($path)
  {
    if(Str::substring($path, 0, Str::length($path) - 1) != "/")
    {
      $path .= "/";
    }
    $this->templatePath2 = $path;
    return $this;
  }

  /**
  * Sets a new layout template.
  *
  * @param string	Template name
  *
  * @return Template
  */
  public function setLayoutTemplate($template = null)
  {
    if(is_null($template))
    {
      $template = Core::getConfig()->maintemplate;
    }
    $this->mainTemplateFile = $template;
    return $this;
  }

  /**
  * Sets a new template file extension.
  *
  * @param string
  *
  * @return Template
  */
  public function setExtension($extension = null)
  {
    if(is_null($extension))
    {
      $extension = Core::getConfig()->templateextension;
    }
    if(Str::substring($extension, 0, 1) != ".")
    {
      $extension = ".".$extension;
    }
    $this->templateExtension = $extension;
    return $this;
  }

  /**
  * Sets the template package.
  *
  * @param string	Package direcotry
  *
  * @return Template
  */
  public function setTemplatePackage($package = null)
  {
    if(is_null($package))
    {
      $package = Core::getConfig()->templatepackage;
    }
    if(Str::substring($package, 0, Str::length($package) - 1) != "/")
    {
      $package .= "/";
    }
    $this->templatePackage = $package;
    return $this;
  }

  /**
  * Returns the template package.
  *
  * @return string
  */
  public function getTemplatePackage()
  {
    if(is_dir($this->templatePath.$this->templatePackage))
    {
      return $this->templatePackage;
    }
    return "standard/";
  }

  /**
  * Adds a message to log.
  *
  * @param string	Message to add
  *
  * @return Template
  */
  public function addLogMessage($message)
  {
    $this->log->push($message);
    return $this;
  }

  /**
  * Adds an HTML header file to layout template.
  *
  * @param string	File to include
  * @param string	File type [optional]
  *
  * @return Template
  */
  public function addHTMLHeaderFile($file, $type = "js")
  {
    $type = strtolower($type);

    if(isset($this->headKeys[$type][$file]))
    {
      return $this;
    }
    $this->headKeys[$type][$file] = 1;

    switch($type)
    {
    case "css":
      $file = RELATIVE_URL."css/".$file;
      $head = "<link rel=\"stylesheet\" type=\"text/css\" href=\"".$file."\" media=\"screen\" />";
      break;
    case "js":
    default:
      $file = RELATIVE_URL."js/".$file;
      $head = "<script type=\"text/javascript\" src=\"".$file."\"></script>";
      break;
    }
    $this->htmlHead->push($head);
    return $this;
  }

  /**
  * Returns an assigned template variable.
  *
  * @param string	Variable name
  * @param mixed		Default value to return
  *
  * @return mixed
  */
  public function get($var, $default = null)
  {
    $default = ($default == "[var]") ? $var : $default;
    return (isset($this->templateVars[$var])) ? $this->templateVars[$var] : $default;
  }

  /**
  * Returns an assigned loop.
  *
  * @param string	Loop index
  *
  * @return mixed
  */
  public function getLoop($loop)
  {
    if($loop[0] != ".")
    {
      return isset($this->loopStack[$loop]) ? $this->loopStack[$loop] : array();
    }
    $parts = explode(".", $loop);
    $i = count($parts)-1;

    // debug_var($parts, "[getLoop] $loop, parts");
    // debug_var($this->runValueStack, "[getLoop] $loop, runValueStack");

    if(isset($this->runValueStack[$i-1][$parts[$i]]))
    {
      // debug_var($this->runValueStack[$i-1][$parts[$i]], "[getLoop] $loop, return");
      return $this->runValueStack[$i-1][$parts[$i]];
    }
    return array();
  }

  public function getLoopVar($name)
  {
    if($name[0] != ".")
    {
      $i = count($this->runValueStack)-1;
      return $this->runValueStack[$i][$name];
    }

    $parts = explode(".", $name);
    $i = count($parts)-1;
    return $this->runValueStack[$i-1][$parts[$i]];
  }

  public function getLoopRow($name = "")
  {
    if($name[0] != ".")
    {
      $i = count($this->runValueStack)-1;
      return $this->runValueStack[$i];
    }

    $parts = explode(".", $name);
    $i = count($parts)-1;
    return $this->runValueStack[$i-1];
  }

  /**
  * Sets the view class.
  *
  * @param object
  *
  * @return Template
  */
  public function setView($view)
  {
    if(is_object($view))
    {
      $this->view = $view;
    }
    return $this;
  }

  /**
  * Returns the view class.
  *
  * @return object
  */
  public function getView()
  {
    return $this->view;
  }
}
?>