<?php
/**
* All file system related functions.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: File.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class File
{
  /**
  * Fetch the extension of a file name.
  *
  * @param string	The file name.
  *
  * @return string 	The file extension.
  */
  public static function getFileExtension($filename)
  {
    return strtolower(Str::substring(strrchr($filename, "."), 1));
  }

  /**
  * Deletes a file.
  *
  * @param string	Path + file name
  *
  * @return boolean	True on success or false on failure
  */
  public static function rmFile($file)
  {
    if(is_dir($file))
    {
      return self::rmDirectory($file);
    }
    if(file_exists($file))
    {
      if(!@unlink($file))
      {
        throw new GenericException("Cannot delete file \"".$file."\".");
        return false;
      }
    }
    else
    {
      throw new GenericException("Cannot delete a non-existing file (\"".$file."\").");
      return false;
    }
    return true;
  }

  /**
  * Deletes a complete direcotory including its contents.
  *
  * @param string	Path
  *
  * @return boolean	True on success or false on failure
  */
  public static function rmDirectory($dir)
  {
    if(is_dir($dir))
    {
      if(self::rmDirectoryContent($dir))
      {
        @rmdir($dir);
        return true;
      }
      return false;
    }
    return false;
  }

  /**
  * Deletes the complete directory content.
  *
  * @param string	Directory path
  *
  * @return boolean	True on success or false on failure
  */
  public static function rmDirectoryContent($dir)
  {
    if(is_dir($dir))
    {
      $dir = (substr($dir, -1) != "/") ? $dir."/" : $dir;
      $openDir = opendir($dir);
      while($file = readdir($openDir))
      {
        if($file == "." || $file == ".." || $file == "Thumbs.db") // Thumbs.db causes bugs on win.
        {
          continue;
        }
        if(!is_dir($dir.$file)) { self::rmFile($dir.$file); }
        else { self::rmDirectory($dir.$file); }
      }
      closedir($openDir);
    }
    return false;
  }

  /**
  * Moves the complete direcotory content into another directory.
  *
  * @param string	Folder to move
  * @param string	Destination folder
  *
  * @return boolean	True on success or false on failure
  */
  public static function mvDirectoryContent($from, $to)
  {
    if(is_dir($from))
    {
      if(!is_dir($to)) { @mkdir($to); }
      $from = (substr($from, -1) != "/") ? $from."/" : $from;
      $to = (substr($to, -1) != "/") ? $to."/" : $to;
      $openDir = opendir($from);
      while($file = readdir($openDir))
      {
        if($file == "." || $file == ".." || $file == "Thumbs.db" || $file == ".svn") // Thumbs.db causes bugs on win.
        {
          continue;
        }
        $entry = $from.$file;
        if(is_dir($entry))
        {
          self::mvDirectory($entry, $to.$file );
          continue;
        }
        copy($entry, $to.$file);
        self::rmFile($entry);
      }
      return true;
    }
    return false;
  }

  /**
  * Copies the complete direcotory content into another directory.
  *
  * @param string	Folder to copy
  * @param string	Destination folder
  *
  * @return boolean	True on success or false on failure
  */
  public static function cpDirectoryContent($from, $to)
  {
    if(is_dir($from))
    {
      if(!is_dir($to)) { @mkdir($to); }
      $from = (substr($from, -1) != "/") ? $from."/" : $from;
      $to = (substr($to, -1) != "/") ? $to."/" : $to;
      $openDir = opendir($from);
      while($file = readdir($openDir))
      {
        if($file == "." || $file == ".." || $file == "Thumbs.db" || $file == ".svn") // Thumbs.db causes bugs on win.
        {
          continue;
        }
        $entry = $from.$file;
        if(is_dir($entry))
        {
          self::cpDirectoryContent($entry, $to.$file );
          continue;
        }
        copy($entry, $to.$file);
      }
      return true;
    }
    return false;
  }

  /**
  * Makes a copy of the file source to destination.
  *
  * @param string	File to copy
  * @param string	Destination path
  *
  * @return boolean	True on success or false on failure
  */
  public static function cpFile($file, $dest)
  {
    if(is_dir($file))
    {
      return self::cpDirectoryContent($file, $dest);
    }
    if(!file_exists($file))
    {
      throw new GenericException("Cannot copy a non-existing file (\"".$file."\").");
      return false;
    }
    if(!is_dir(dirname($dest)))
    {
      throw new GenericException("Copy destination is not writable (\"".$file."\").");
      return false;
    }
    if(!copy($file, $dest))
    {
      throw new GenericException("Unable to copy \"".$file."\" to \"".$dest."\".");
      return false;
    }
    return true;
  }

  /**
  * Returns file size.
  *
  * @param string	Path to file
  * @param boolean	Format size or return raw byte value?
  *
  * @return mixed	File size
  */
  public static function getFileSize($file, $format = true)
  {
    $size = (is_dir($file)) ? self::getDirectorySize($file, false) : filesize($file);
    return ($format) ? self::bytesToString($size) : $size;
  }

  /**
  * Returns directory size.
  *
  * @param string	Path to directory
  * @param boolean	Format size or return raw byte value?
  *
  * @return mixed	Directory size
  */
  public static function getDirectorySize($dir, $format = true)
  {
    $size = 0;
    $handle = opendir($dir);
    $dir .= "/";
    while($file = readdir($handle))
    {
      if($file != "." && $file != ".." && $file != "Thumbs.db" && $file != ".svn" && $file != ".htaccess")
      {
        $size += (is_dir($dir.$file)) ? self::getDirectorySize($dir.$file, false) : filesize($dir.$file);
      }
    }
    closedir($handle);
    return ($format) ? self::bytesToString($size) : $size;
  }

  /**
  * Converts byte number into readable string.
  *
  * @param integer	Bytes to convert
  *
  * @return string
  */
  public static function bytesToString($bytes)
  {
    if($bytes == 0) { return number_format($bytes, 2)." Byte"; }
    $s = array("Byte", "Kb", "MB", "GB", "TB", "PB");
    $e = floor(log($bytes)/log(1024));
    return sprintf("%.2f ".$s[$e], ($bytes/pow(1024, floor($e))));
  }
}
?>