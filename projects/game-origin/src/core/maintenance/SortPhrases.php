<?php
/**
* Function to resort phrases.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
*/

function sortPhrases()
{
//  try
//  {
    Core::getDB()->query("DROP TABLE IF EXISTS ".PREFIX."buffer");
//  }
//  catch(Exception $e) { $e->printError(); }

//  try
//  {
    Core::getDB()->query("CREATE TABLE ".PREFIX."buffer " .
      "(phraseid int(10) unsigned NOT NULL auto_increment, " .
      "languageid int(4) unsigned NOT NULL, " .
      "phrasegroupid int(4) unsigned NOT NULL, " .
      "title varbinary(128) NOT NULL, " .
      "content varbinary(10000) NOT NULL, " .
      "PRIMARY KEY (phraseid), " .
      "KEY languageid (languageid,phrasegroupid), " .
      "KEY phrasegroupid (phrasegroupid)) ENGINE=MyISAM DEFAULT CHARSET=binary AUTO_INCREMENT=1");
//  }
//  catch(Exception $e) { $e->printError(); }

//  try
//  {
    Core::getDB()->query("INSERT INTO ".PREFIX."buffer (languageid, phrasegroupid, title, content) SELECT languageid, phrasegroupid, title, content FROM ".PREFIX."phrases ORDER BY languageid ASC, phrasegroupid ASC, title ASC");
//  }
//  catch(Exception $e) { $e->printError(); }

//  try
//  {
    Core::getDB()->query("TRUNCATE TABLE ".PREFIX."phrases");
//  }
//  catch(Exception $e) { $e->printError(); }

//  try
//  {
    Core::getDB()->query("INSERT INTO ".PREFIX."phrases (phraseid, languageid, phrasegroupid, title, content) SELECT phraseid, languageid, phrasegroupid, title, content FROM ".PREFIX."buffer ORDER BY phraseid ASC");
//  }
//  catch(Exception $e) { $e->printError(); }

//  try
//  {
    Core::getDB()->query("DROP TABLE IF EXISTS ".PREFIX."buffer");
//  }
//  catch(Exception $e) { $e->printError(); }
}
?>
