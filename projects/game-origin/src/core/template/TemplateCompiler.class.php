<?php
/**
* Template compiler. Generates PHP code of an template.
* Note: There is still no error reporting available.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: TemplateCompiler.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class TemplateCompiler extends Cache
{
  /**
  * Source template code.
  *
  * @var string
  */
  protected $sourceTemplate = "";

  /**
  * Template file name.
  *
  * @var string
  */
  protected $template = "";

  /**
  * Regular expression patterns.
  *
  * @var array
  */
  protected $patterns = array();

  /**
  * The compiled template content.
  *
  * @var String
  */
  protected $compiledTemplate = null;

  /**
  * Constructor: Set up compiler.
  *
  * @param string	Template name
  *
  * @return void
  */
  public function __construct($template)
  {
    $this->template = Str::reverse_strrchr(basename($template), '.', true);
    $this->template = Str::substring($this->template, 0, Str::length($this->template) - 1);
    $this->sourceTemplate = file_get_contents($template);
    $this->buildPatterns()->compile();
//    try {
    	parent::putCacheContent(Core::getCache()->getTemplatePath($this->template), $this->compiledTemplate->get());
//    }
//    catch(Exception $e) { $e->printError(); }
    return;
  }

  /**
  * Builds the compiling patterns.
  *
  * @return TemplateCompiler
  */
  protected function buildPatterns()
  {
    $this->patterns["var"][] = "/\{var}([^\"]+)\{\/var}/siU";
    $this->patterns["var"][] = "/\{var=([^\"]+)\}/siU";
    $this->patterns["link"] = "/\{link\[([^\"]+)]}(.*)\{\/link}/siU";
    $this->patterns["phrase"][] = "/\{lang}([^\"]+)\{\/lang}/siU";
    $this->patterns["phrase"][] = "/\{lang=([^\"]+)\}/siU";
    $this->patterns["config"][] = "/\{config}([^\"]+)\{\/config}/siU";
    $this->patterns["config"][] = "/\{config=([^\"]+)\}/siU";
    $this->patterns["user"][] = "/\{user}([^\"]+)\{\/user}/siU";
    $this->patterns["user"][] = "/\{user=([^\"]+)\}/siU";
    $this->patterns["request"] = "/\{request\[([^\"]+)\]\}([^\"]+)\{\/request\}/siU";
    $this->patterns["const"][] = "/\{const}([^\"]+)\{\/const}/siU";
    $this->patterns["const"][] = "/\{const=([^\"]+)\}/siU";
    $this->patterns["permission"] = "/\{perm\[([^\"]+)\]\}(.*)\{\/perm\}/siU";
    $this->patterns["include"][] = "/\{include}(.*)\{\/include}/siU";
    $this->patterns["include"][] = "/\{include=(.*)\}/siU";
    $this->patterns["image"] = "/\{image\[([^\"]+)]}([^\"]+)\{\/image}/siU";
    $this->patterns["assignment"] = "/\{\@([^\"]+)}/siU";
    $this->patterns["php_time"] = "/\{PHPTime}/siU";
    $this->patterns["sql_queries"] = "/\{SQLQueries}/siU";
    $this->patterns["db_time"] = "/\{DBTime}/siU";
    $this->patterns["time"][] = "/\{time}(.*)\{\/time}/siU";
    $this->patterns["time"][] = "/\{time=(.*)\}/siU";

    $this->patterns["if"] = "/\{if\[(.*)]}/siU";
    $this->patterns["endif"] = "/\{\/if}/siU";
    $this->patterns["else"] = "/\{else}/siU";
    $this->patterns["elseif"] = "/\{else if\[(.*)]}/siU";

    $this->patterns["loopvar"][] = "/\{loop}(\.[^\"]*)\{\/loop}/siU";
    $this->patterns["loopvar"][] = "/\{loop=(\.[^\"]*)\}/siU";
    $this->patterns["loopvarfast"][] = "/\{loop}([^\.][^\"]*)\{\/loop}/siU";
    $this->patterns["loopvarfast"][] = "/\{loop=([^\.][^\"]*)\}/siU";
    $this->patterns["while"] = "/\{while\[([^\"]+)]}(.*)\{\/while}/siU";

    $this->patterns["foreach"] = "/\{foreach\[([^\"]+)\]}(.*)\{\/foreach}/siU";
    $this->patterns["foreach2"] = "/\{foreach2\[(\.[^\.][^\"]*)\]}(.*)\{\/foreach2}/siU";
    $this->patterns["foreach3"] = "/\{foreach3\[(\.\.[^\.][^\"]*)\]}(.*)\{\/foreach3}/siU";
    $this->patterns["foreach4"] = "/\{foreach4\[(\.\.\.[^\.][^\"]*)\]}(.*)\{\/foreach4}/siU";
    $this->patterns["totalloopvars"] = "/\{\~count}/siU";

    $this->patterns["dphptags"] = "/\?><\?php/siU";
    $this->patterns["hooks"][] = "/\{hook}([^\"]+)\{\/hook}/siU";
    $this->patterns["hooks"][] = "/\{hook=([^\"]+)\}/siU";
    return $this;
  }

  /**
  * Compiles source template code into PHP code.
  *
  * @return TemplateCompiler
  */
  protected function compile()
  {
    $this->compiledTemplate = new String($this->sourceTemplate);

    // remove utf-8 prefix
    $this->compiledTemplate
      ->regEx("#^".preg_quote("\xEF\xBB\xBF", "#")."\s*#is", "");

    // Compile variables
    $this->compiledTemplate
      ->regEx($this->patterns["var"], "\$this->get(\"$1\", false)")
      // Compile links {link[varname]}"index.php/Main"{/link}
      ->regEx($this->patterns["link"], "<?php echo Link::get(\\2, Core::getLanguage()->get(\"\\1\", \$this)); ?>")
      // Compile language variables {lang}varname{/lang}
      ->regEx($this->patterns["phrase"], "<?php echo Core::getLanguage()->get(\"\\1\", \$this); ?>")
      // Compile config variables {config}varname{/config}
      ->regEx($this->patterns["config"], "<?php echo Core::getOptions()->get(\"\\1\"); ?>")
      // Compile session variables {user}varname{/user}
      ->regEx($this->patterns["user"], "<?php echo Core::getUser()->get(\"\\1\"); ?>")
      // Compile request variables {request[get]}varname{/request}{request[post]}varname{/request}{request[cookie]}varname{/request}
      ->regEx($this->patterns["request"], "<?php echo Core::getRequest()->get(\"\\1\", \"\\2\"); ?>")
      // Compile constants {const}CONSTANT{/const}
      ->regEx($this->patterns["const"], "<?php echo \\1; ?>")
      // Parse permission expression {perm[CAN_READ_THIS]}print this{/perm}
      ->regEx($this->patterns["permission"], "{if[Core::getUser()->ifPermissions(\"\\1\")]}\\2{/if}")
      // Compile includes {include}"templatename"{/include}
      ->regEx($this->patterns["include"], "<?php \$this->includeTemplate(\\1); ?>")
      // Compile images {image[title]}path/pic.jpg{/image}
      ->regEx($this->patterns["image"], "<?php echo Image::getImage(\"\\2\", Core::getLanguage()->getItem(\"\\1\", \$this)); ?>")
      // Compile generation times.
      ->regEx($this->patterns["php_time"], "<?php echo Core::getTimer()->getTime(); ?>")
      ->regEx($this->patterns["db_time"], "<?php echo Core::getDB()->getQueryTime(); ?>")
      ->regEx($this->patterns["sql_queries"], "<?php echo Core::getDB()->getQueryNumber(); ?>")
      // Compile hook tags
      ->regEx($this->patterns["hooks"], "<?php Hook::event(\"\\1\", array(\$this)); ?>")
      // Compile time designations.
      ->regEx($this->patterns["time"], "<?php echo Date::timeToString(3, -1, \"\\1\", false); ?>");

    $this->compileIfTags()->compileLoops();

    $this	->compiledTemplate
      // Compile wildcards {@assignment}
      ->regEx($this->patterns["assignment"], "<?php echo \$this->get(\"\\1\"); ?>")
      // Remove useless double php tags.
      ->regEx($this->patterns["dphptags"], "");

    // Hook::event("COMPILE_TEMPLATE", array($this, &$this->compiledTemplate));

    $this	->compiledTemplate
      ->pop(parent::setCacheFileHeader("Template Cache File")."?>\r")
      ->push("\r\r<?php // Cache-Generator finished ?>");
    return $this;
  }

  /**
  * Compiles if-else tags into PHP code.
  *
  * @return TemplateCompiler
  */
  protected function compileIfTags()
  {
    // Fetch complexes if else tags like {if[term]}print this{else if[term]}print that{else if[term]}print this [...]{else}or print this{/if}
    $this	->compiledTemplate
      ->regEx($this->patterns["if"], "<?php if($1) { ?>")
      ->regEx($this->patterns["endif"], "<?php } ?>")
      ->regEx($this->patterns["else"], "<?php } else { ?>")
      ->regEx($this->patterns["elseif"], "<?php } else if($1) { ?>");
    return $this;
  }

  /**
  * Compiles loops {while[resource]}Print this{/while} or {foreach[array]}Print this{/while}.
  *
  * @return TemplateCompiler
  */
  protected function compileLoops()
  {
    $this	->compiledTemplate
      // While loops (Specially for multiple database queries).
      ->regEx($this->patterns["while"], "<?php while(\$row = Core::getDB()->fetch(\$this->getLoop(\"$1\"))){ ?> $2 <?php } ?>")
      // Foreach loops (Specially for arrays).
      ->regEx($this->patterns["foreach"], "<?php \$cur_loop = \$this->getLoop(\"$1\"); \$count = count(\$cur_loop); foreach(\$cur_loop as \$key => \$row) { array_push(\$this->runValueStack, \$row); ?> $2 <?php array_pop(\$this->runValueStack); } ?>")
      ->regEx($this->patterns["foreach2"], "<?php array_push(\$this->runTempStack, array(\$count, \$row)); \$cur_loop = \$this->getLoop(\"$1\"); \$count = count(\$cur_loop); foreach(\$cur_loop as \$key => \$row) { array_push(\$this->runValueStack, \$row); ?> $2 <?php array_pop(\$this->runValueStack); } list(\$count, \$row) = array_pop(\$this->runTempStack);  ?>")
      ->regEx($this->patterns["foreach3"], "<?php array_push(\$this->runTempStack, array(\$count, \$row)); \$cur_loop = \$this->getLoop(\"$1\"); \$count = count(\$cur_loop); foreach(\$cur_loop as \$key => \$row) { array_push(\$this->runValueStack, \$row); ?> $2 <?php array_pop(\$this->runValueStack); } list(\$count, \$row) = array_pop(\$this->runTempStack);  ?>")
      ->regEx($this->patterns["foreach4"], "<?php array_push(\$this->runTempStack, array(\$count, \$row)); \$cur_loop = \$this->getLoop(\"$1\"); \$count = count(\$cur_loop); foreach(\$cur_loop as \$key => \$row) { array_push(\$this->runValueStack, \$row); ?> $2 <?php array_pop(\$this->runValueStack); } list(\$count, \$row) = array_pop(\$this->runTempStack);  ?>")
      // Variables within loops {while[resource]}{loop}column1{/loop}{/while}
      ->regEx($this->patterns["loopvar"], "<?php echo \$this->getLoopVar(\"$1\"); ?>")
      ->regEx($this->patterns["loopvarfast"], "<?php echo \$row[\"$1\"]; ?>")
      // Total Number of array elements.
      ->regEx($this->patterns["totalloopvars"], "<?php echo \$count; ?>");

    /*
    $old_str = $this->compiledTemplate->get();
    for(;;)
    {
      $this->compiledTemplate
        // Foreach loops (Specially for arrays).
        ->regEx($this->patterns["foreach"], "<?php \$cur_loop = \$this->getLoop(\"$1\"); \$count = count(\$cur_loop); foreach(\$cur_loop as \$key => \$row) { array_push(\$this->runValueStack, \$row); ?> $2 <?php array_pop(\$this->runValueStack); } ?>");
      $str = $this->compiledTemplate->get();
      if($old_str == $str)
      {
        break;
      }
      $old_str = $str;
    }
    */
    return $this;
  }
}
?>