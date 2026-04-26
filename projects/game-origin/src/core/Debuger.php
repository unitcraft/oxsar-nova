<?php
/**
* Oxsar http://oxsar.ru
*
* 
*/

$GLOBALS['debug_params'] = array('outputFormat' => 'html');

/**
* Print_r convenience function, which prints out <PRE> tags around
* the output of given array. Similar to debug().
*
* @see	debug()
* @param array $var Variable to print out
* @param boolean $showFrom If set to true, the method prints from where the function was called
* @link http://book.cakephp.org/view/707/pr
*/
function pr($var)
{
  // if (Configure::read() > 0)
  {
    echo '<pre>';
    print_r($var);
    echo '</pre>';
  }
}

/**
* Convenience method for htmlspecialchars.
*
* @param string $text Text to wrap through htmlspecialchars
* @param string $charset Character set to use when escaping.  Defaults to config value in 'App.encoding' or 'UTF-8'
* @return string Wrapped text
* @link http://book.cakephp.org/view/703/h
*/
function h($text, $charset = null)
{
  if (is_array($text)) {
    return array_map('h', $text);
  }
  /* if (empty($charset)) {
  $charset = Configure::read('App.encoding');
  } */
  if (empty($charset)) {
    $charset = 'UTF-8';
  }
  return htmlspecialchars($text, ENT_QUOTES, $charset);
}

function debug_var($v, $text, $recursion = 5)
{
  echo "<div style='padding-left: 170px'>";
  echo "$text";
  pr($v);
  echo "</div>";
  // pr(debug_export_var($v, $recursion));
  flush();
}

function debug_trace($options = array())
{
  pr(debug_export_trace($options));
}

/**
* Outputs a stack trace based on the supplied options.
*
* @param array $options Format for outputting stack trace
* @return string Formatted stack trace
* @access public
* @static
* @link http://book.cakephp.org/view/460/Using-the-Debugger-Class
*/
function debug_export_trace($options = array())
{
  $options = array_merge(array(
    'depth'		=> 999,
    'format'	=> '',
    'args'		=> false,
    'start'		=> 0,
    'scope'		=> null,
    'exclude'	=> null
    ),
    $options
    );

  $backtrace = debug_backtrace();
  $back = array();
  $count = count($backtrace);

  for ($i = $options['start']; $i < $count && $i < $options['depth']; $i++) {
    $trace = array_merge(
      array(
      'file' => '[internal]',
      'line' => '??'
      ),
      $backtrace[$i]
    );

    if (isset($backtrace[$i + 1])) {
      $next = array_merge(
        array(
        'line'		=> '??',
        'file'		=> '[internal]',
        'class'		=> null,
        'function'	=> '[main]'
        ),
        $backtrace[$i + 1]
      );
      $function = $next['function'];

      if (!empty($next['class'])) {
        $function = $next['class'] . '::' . $function . '(';
        if ($options['args'] && isset($next['args'])) {
          $args = array();
          foreach ($next['args'] as $arg) {
            $args[] = debug_export_var($arg);
          }
          $function .= join(', ', $args);
        }
        $function .= ')';
      }
    } else {
      $function = '[main]';
    }
    if (in_array($function, array('call_user_func_array', 'trigger_error'))) {
      continue;
    }
    if ($options['format'] == 'points' && $trace['file'] != '[internal]') {
      $back[] = array('file' => $trace['file'], 'line' => $trace['line']);
    } elseif (empty($options['format'])) {
      $back[] = $function . ' - ' . debug_trim_path($trace['file']) . ', line ' . $trace['line'];
    } else {
      $back[] = $trace;
    }
  }

  if ($options['format'] == 'array' || $options['format'] == 'points') {
    return $back;
  }
  return join("\n", $back);
}

/**
* Converts a variable to a string for debug output.
*
* @param string $var Variable to convert
* @return string Variable as a formatted string
* @access public
* @static
* @link http://book.cakephp.org/view/460/Using-the-Debugger-Class
*/
function debug_export_var($var, $recursion = 0)
{
  switch (strtolower(gettype($var)))
  {
  case 'boolean':
    return ($var) ? 'true' : 'false';

  case 'integer':
  case 'double':
    return $var;

  case 'string':
    if (trim($var) == "") {
      return '""';
    }
    return '"' . h($var) . '"';

  case 'object':
    return get_class($var) . "\n" . debug_export_object($var);

  case 'array':
    $out = "array(";
    $vars = array();
    foreach ($var as $key => $val) {
      if ($recursion >= 0) {
        if (is_numeric($key)) {
          $vars[] = "\n\t" . debug_export_var($val, $recursion - 1);
        } else {
          $vars[] = "\n\t" .debug_export_var($key, $recursion - 1)
            . ' => ' . debug_export_var($val, $recursion - 1);
        }
      }
    }
    $n = null;
    if (count($vars) > 0) {
      $n = "\n";
    }
    return $out . join(",", $vars) . "{$n})";

  case 'resource':
    return strtolower(gettype($var));

  case 'null':
    return 'null';
  }
  return 'null';
}

/**
* Handles object to string conversion.
*
* @param string $var Object to convert
* @return string
* @access private
* @see Debugger:exportVar()
*/
function debug_export_object($var)
{
  $out = array();

  if (is_object($var))
  {
    $className = get_class($var);
    $objectVars = get_object_vars($var);

    foreach ($objectVars as $key => $value)
    {
      if (is_object($value)) {
        $value = get_class($value) . ' object';
      } elseif (is_array($value)) {
        $value = 'array';
      } elseif ($value === null) {
        $value = 'NULL';
      } elseif (in_array(gettype($value), array('boolean', 'integer', 'double', 'string', 'array', 'resource'))) {
        $value = debug_export_var($value);
      }
      $out[] = "$className::$$key = " . $value;
    }
  }
  return join("\n", $out);
}

/**
* Shortens file paths by replacing the application base path with 'APP', and the CakePHP core
* path with 'CORE'.
*
* @param string $path Path to shorten
* @return string Normalized path
* @access public
* @static
*/
function debug_trim_path($path)
{
  if(defined('APP_ROOT_DIR') && strpos($path, APP_ROOT_DIR) === 0)
  {
    return 'APP/' . substr($path, strlen(APP_ROOT_DIR), 1000);
  }

  return $path;
}

/**
* Overrides PHP's default error handling.
*
* @param integer $code Code of error
* @param string $description Error description
* @param string $file File on which error occurred
* @param integer $line Line that triggered the error
* @param array $context Context
* @return boolean true if error was handled
* @access public
*/
function debug_error_handler($code, $description, $file = null, $line = null, $context = null)
{
  // echo "[debug_error_handler] <br />";

  if (error_reporting() == 0 || $code === 2048 || $code === 8192)
  {
    return;
  }

  if (empty($file)) {
    $file = '[internal]';
  }
  if (empty($line)) {
    $line = '??';
  }
  $file = debug_trim_path($file);

  $info = compact('code', 'description', 'file', 'line');
  /* if (!in_array($info, $_this->errors)) {
  $_this->errors[] = $info;
  } else {
  return;
  } */

  $level = LOG_DEBUG;
  switch ($code) {
    case E_PARSE:
    case E_ERROR:
    case E_CORE_ERROR:
    case E_COMPILE_ERROR:
    case E_USER_ERROR:
      $error = 'Fatal Error';
      $level = LOG_ERROR;
      break;
    case E_WARNING:
    case E_USER_WARNING:
    case E_COMPILE_WARNING:
    case E_RECOVERABLE_ERROR:
      $error = 'Warning';
      $level = LOG_WARNING;
      break;
    case E_NOTICE:
    case E_USER_NOTICE:
      $error = 'Notice';
      $level = LOG_NOTICE;
      break;
    default:
      return false;
      break;
  }

  $helpCode = null;
  /* if (!empty($_this->helpPath) && preg_match('/.*\[([0-9]+)\]$/', $description, $codes)) {
  if (isset($codes[1])) {
  $helpCode = $codes[1];
  $description = trim(preg_replace('/\[[0-9]+\]$/', '', $description));
  }
  } */

  debug_output_internal($level, $error, $code, $helpCode, $description, $file, $line, $context);

  /*
  if (Configure::read('log')) {
  CakeLog::write($level, "{$error} ({$code}): {$description} in [{$file}, line {$line}]");
  }
  */

  if ($error == 'Fatal Error') {
    die();
  }
  return true;
}

function debug_exception_handler($exception)
{
  $file = $exception->getFile();
  $line = $exception->getLine();

  if (empty($file))
  {
    $file = '[internal]';
  }
  if (empty($line))
  {
    $line = '??';
  }
  $file = debug_trim_path($file);

  debug_output_internal(LOG_ERROR, 'Exception', $exception->getCode(), null, $exception->getMessage(), $file, $line, null);
  die();
}

/**
* Grabs an debug_excerpt from a file and highlights a given line of code
*
* @param string $file Absolute path to a PHP file
* @param integer $line Line number to highlight
* @param integer $context Number of lines of context to extract above and below $line
* @return array Set of lines highlighted
* @access public
* @static
* @link http://book.cakephp.org/view/460/Using-the-Debugger-Class
*/
function debug_excerpt($file, $line, $context = 2)
{
  $data = $lines = array();
  if (!file_exists($file)) {
    return array();
  }
  $data = @explode("\n", file_get_contents($file));

  if (empty($data) || !isset($data[$line])) {
    return;
  }
  for ($i = $line - ($context + 1); $i < $line + $context; $i++) {
    if (!isset($data[$i])) {
      continue;
    }
    $string = str_replace(array("\r\n", "\n"), "", highlight_string($data[$i], true));
    if ($i == $line) {
      $lines[] = '<span class="code-highlight">' . $string . '</span>';
    } else {
      $lines[] = $string;
    }
  }
  return $lines;
}

/**
* Handles object conversion to debug string.
*
* @param string $var Object to convert
* @access private
*/
function debug_output_internal($level, $error, $code, $helpCode, $description, $file, $line, $kontext)
{
  global $debug_params;

  $files = debug_export_trace(array('start' => 2, 'format' => 'points'));
  $listing = debug_excerpt($files[0]['file'], $files[0]['line'] - 1, 1);
  $trace = debug_export_trace(array('start' => 2, 'depth' => '20'));
  $context = array();

  foreach ((array)$kontext as $var => $value) {
    $context[] = "\${$var}\t=\t" . debug_export_var($value, 1);
  }

  switch ($debug_params['outputFormat'])
  {
    /*
    default:
    case 'js':
    $link = "document.getElementById(\"CakeStackTrace" . count($this->errors) . "\").style.display = (document.getElementById(\"CakeStackTrace" . count($this->errors) . "\").style.display == \"none\" ? \"\" : \"none\")";
    $out = "<a href='javascript:void(0);' onclick='{$link}'><b>{$error}</b> ({$code})</a>: {$description} [<b>{$file}</b>, line <b>{$line}</b>]";
    if (Configure::read() > 0) {
    debug($out, false, false);
    echo '<div id="CakeStackTrace' . count($this->errors) . '" class="cake-stack-trace" style="display: none;">';
    $link = "document.getElementById(\"CakeErrorCode" . count($this->errors) . "\").style.display = (document.getElementById(\"CakeErrorCode" . count($this->errors) . "\").style.display == \"none\" ? \"\" : \"none\")";
    echo "<a href='javascript:void(0);' onclick='{$link}'>Code</a>";

    if (!empty($context)) {
    $link = "document.getElementById(\"CakeErrorContext" . count($this->errors) . "\").style.display = (document.getElementById(\"CakeErrorContext" . count($this->errors) . "\").style.display == \"none\" ? \"\" : \"none\")";
    echo " | <a href='javascript:void(0);' onclick='{$link}'>Context</a>";

    if (!empty($helpCode)) {
    echo " | <a href='{$this->helpPath}{$helpCode}' target='_blank'>Help</a>";
    }

    echo "<pre id=\"CakeErrorContext" . count($this->errors) . "\" class=\"cake-context\" style=\"display: none;\">";
    echo implode("\n", $context);
    echo "</pre>";
    }

    if (!empty($listing)) {
    echo "<div id=\"CakeErrorCode" . count($this->errors) . "\" class=\"cake-code-dump\" style=\"display: none;\">";
    pr(implode("\n", $listing) . "\n", false);
    echo '</div>';
    }
    pr($trace, false);
    echo '</div>';
    }
    break;
    */

  default:
  case 'html':
    echo "<pre class=\"cake-debug\"><b>{$error}</b> ({$code}) : {$description} [<b>{$file}</b>, line <b>{$line}]</b></pre>";
    if (!empty($context)) {
      echo "Context:\n" .implode("\n", $context) . "\n";
    }
    echo "<pre class=\"cake-debug context\"><b>Context</b> <p>" . implode("\n", $context) . "</p></pre>";
    echo "<pre class=\"cake-debug trace\"><b>Trace</b> <p>" . $trace. "</p></pre>";
    break;

  case 'text':
  case 'txt':
    echo "{$error}: {$code} :: {$description} on line {$line} of {$file}\n";
    if (!empty($context)) {
      echo "Context:\n" .implode("\n", $context) . "\n";
    }
    echo "Trace:\n" . $trace;
    break;

    /*
    case 'log':
    $this->log(compact('error', 'code', 'description', 'line', 'file', 'context', 'trace'));
    break;
    case false:
    $this->__data[] = compact('error', 'code', 'description', 'line', 'file', 'context', 'trace');
    break;
    */
  }
}

if($GLOBALS["RUN_YII"] != 1)
{
	if (!defined('DISABLE_DEBUGER_ERROR_HANDLING'))
	{
	  set_error_handler('debug_error_handler', ERROR_REPORTING_TYPE);
	  set_exception_handler('debug_exception_handler');
	  // echo "set_error_handler <br />";
	  // debug_output_internal(123);
	}
}

?>