<?php
/**
* Link-related functions. Mainly used to generate HTML Links.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Link.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Link
{
	/**
	* CSS class for external links.
	*
	* @var string
	*/
	const CSS_EXTERNAL_URL = "external";

	/**
	* CSS class for standard links.
	*
	* @var string
	*/
	const CSS_NORMAL_URL = "link";

	/**
	* Enable this to append host and path automatically
	* to a link url.
	*
	* @var boolean
	*/
	const APPEND_HOST_PATH = true;

	/**
	* Appends the language code to each URL.
	*
	* @var boolean
	*/
	const APPEND_LANG_TO_URL = false;

	/**
	* Set required data for link.
	*
	* @param string	URL to link
	* @param string	Link name
	* @param string	Additional title
	* @param string	Particular css class
	* @param string	Additional attachment for link
	* @param boolean	Activate URL rewrite
	*
	* @return string	HTML link
	*/
	public static function get($url, $name, $title = "", $cssClass = "", $attachment = "", $appendSession = false, $rewrite = true, $refdir = true, $appendSNparams = false, $sn_name = 'ODNOKLASSNIKI', $auto_social = true)
	{
		$is_console = defined('YII_CONSOLE') && YII_CONSOLE;
		if(Str::length($cssClass) > 0) { /*$cssClass = $cssClass;*/ }
		else
		{
			if(self::isExternal($url))
			{
				$cssClass = self::CSS_EXTERNAL_URL;
			}
			else
			{
				$cssClass = self::CSS_NORMAL_URL;
			}
		}
		if(Str::length($attachment) > 0)
		{
			$attachment = " ".$attachment;
		}

		if(self::isExternal($url) && $refdir)
		{
			$link = "<a href=\"javascript:void(0);\" onclick=\"window.open('".RELATIVE_URL."refdir.php?url=".$url."');\" title=\"".$title."\" class=\"".$cssClass."\"".$attachment.">".$name."</a>";
		}
		else
		{
			if( $is_console )
			{
				$session = 1;
			}
			else
			{
				$session = Core::getRequest()->getGET("sid");
			}
			
			if(!COOKIE_SESSION && $appendSession && $session)
			{
//				if(Str::inString("?", $url))
//				{
//					$url .= "?sid=".$session;
//				}
//				else
//				{
//					$url .= "&amp;sid=".$session;
//				}
			}
			if($url != "#")
			{
				if($auto_social)
				{
					$url = socialUrl($url);
				}
				/*
				if( defined('SN')
				{
					$suf = '';
					if( strpos($url, $suf) === false )
					{
						if( strpos($url, '?') === false )
						{
							$url .= '';
						}
						else
						{
							$url .= '';
						}
					}
				}
				*/
				
				if( $is_console )
				{
					$lang = 1;
				}
				else
				{
					$lang = (Core::getRequest()->getPOST("lang")) ? Core::getRequest()->getPOST("lang") : Core::getRequest()->getGET("lang");
				}
				
				if(self::APPEND_LANG_TO_URL && $lang)
				{
					if(Str::inString("?", $url))
					{
						$url = $url."?lang=".$lang;
					}
					else
					{
						$url = $url."&amp;lang=".$lang;
					}
				}
				
				if( defined("FORCE_REWRITE") && FORCE_REWRITE && $rewrite && !defined('SN') )
				{
					$url = self::normalizeURL($url);
				}
				elseif( defined("FORCE_REWRITE") && FORCE_REWRITE && $rewrite && defined('SN') )
				{
					$url = self::normalizeSNURL($url);
				}
				else if(!Link::isExternal($url) && !preg_match("~^(".preg_quote(FULL_URL,"~")."|".preg_quote(RELATIVE_URL,"~").")~si", $url) && self::APPEND_HOST_PATH)
				{
					$url = RELATIVE_URL.$url;
				}
			}
			$link = "<a href=\"".$url."\" title=\"".$title."\" class=\"".$cssClass."\"".$attachment.">".$name."</a>";
		}
		return $link;
	}

	/**
	* Check whether url is beyond current site.
	*
	* @param string	URL to check
	*
	* @return boolean
	*/
	public static function isExternal($url)
	{
		if($url == "#")
		{
			return false;
		}
		if(preg_match("#^((http(s)?|ftp|news)://)+[a-z\d\.@_-]*[a-z\d@_-]+\.([a-z]{2}|biz|com|gov|info|int|museum|name|net|org)#i", $url) || preg_match("#^file:///#i", $url))
		{
			return true;
		}
		return false;
	}

	/**
	* Encode URL according to RFC1738 (spaces will be replaced with _).
	*
	* @param string	URL to validate
	*
	* @return string	Validated URL
	*/
	public static function validateURL($url)
	{
		$url = Str::replace(" ", "_", $url);
		$url = rawurlencode($url);
		return $url;
	}

	/**
	* Normalize the URL into readable string for the Rewrite-Engine.
	*
	* @param string	URL to normalize
	*
	* @return string	Normalized URL
	*/
	public static function normalizeURL($url)
	{
		if( defined('SN') )
		{
			return RELATIVE_URL.$url;
		}
		if(strpos($url, "?") > 0)
		{
			$url = preg_replace("/\?(.*?)=/i", "/$1:", $url); // Replace ?arg= with /arg:
			$url = preg_replace("/\&amp;(.*?)=/i", "/$1:", $url); // Replace &amp;arg= with /arg:
			$url = preg_replace("/\&(.*?)=/i", "/$1:", $url); // Replace &arg= with /arg:

			// Now remove useless arg names.
			$parsedURL = parse_url($url);

		$script_name = "/".substr(REQUEST_DIR, 0, -1); // $_SERVER["SCRIPT_NAME"]
			$path = Str::substring(Str::replace($script_name, "", $parsedURL["path"]), 1);
			$splitted = explode("/", $path);
			$size = count($splitted);
			for($i = 0; $i < $size; $i++)
			{
				if(strpos($splitted[$i], ":"))
				{
					$splitted[$i] = explode(":", $splitted[$i]);
					$levelNames = explode(",", REQUEST_LEVEL_NAMES);
					if(Str::compare($splitted[$i][0], $levelNames[$i], true)) { $url = Str::replace($splitted[$i][0].":", "", $url); }
				}
			}
		}
		return RELATIVE_URL.$url;
	}
	
	public static function normalizeSNURL($url)
	{
		if( defined('SN') )
		{
			return RELATIVE_URL.$url;
		}
	}
}
?>