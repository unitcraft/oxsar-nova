<?php
/**
* Use this function to ban several IP blocks at once.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
*
* @param string IP Address (e.g.: 127.0.0.1).
* @param string Subnet mask (e.g.: 16).
* @param string Reason for ban.
*/

function banIPClass($ip, $subnet, $reason = "", $time = 0)
{
  if($subnet < 16 || $subnet > 32)
  {
    throw new GenericException("Subnet mask is invalid.", __FILE__, __LINE__);
  }
  if($time == 0) { $time = time() + 31536000; }
  $ip = explode(".", $ip);
  $binIp[0] = validateBin(decbin($ip[0]), 8);
  $binIp[1] = validateBin(decbin($ip[1]), 8);
  $binIp[2] = validateBin(decbin($ip[2]), 8);
  $binIp[3] = validateBin(decbin($ip[3]), 8);

  $binSubnet = "";
  for($i = 1; $i <= 32; $i++)
  {
    if($i <= $subnet) { $binSubnet .= "1"; }
    else { $binSubnet .= "0"; }
  }

  $binSubnets[0] = Str::substring($binSubnet, 0, 8);
  $binSubnets[1] = Str::substring($binSubnet, 8, 8);
  $binSubnets[2] = Str::substring($binSubnet, 16, 8);
  $binSubnets[3] = Str::substring($binSubnet, 24, 8);

  $network[0] = bindec($binIp[0] & $binSubnets[0]);
  $network[1] = bindec($binIp[1] & $binSubnets[1]);
  $network[2] = bindec($binIp[2] & $binSubnets[2]);
  $network[3] = bindec($binIp[3] & $binSubnets[3]);

  if(bindec($binSubnets[0]) == 255 && bindec($binSubnets[1]) == 255 && bindec($binSubnets[2]) == 255)
  {
    for($i = $network[3] + 1; $i < 255; $i++)
    {
      $banIP[$i] = (string)$network[0].".".(string)$network[1].".".(string)$network[2].".".(string)$i;
    }
  }
  else if(bindec($binSubnets[0]) == 255 && bindec($binSubnets[1]) == 255)
  {
    $m = 0;
    for($i = $network[2] + 1; $i < 255; $i++)
    {
      for($n = 1; $n < 255; $n++)
      {
        $banIP[$m] = $network[0].".".$network[1].".".$i.".".$n;
        $m++;
      }
    }
  }

  $att = array("ipaddress", "reason", "timebegin", "timeend");
  if(count($banIP) > 0) { header('content-type: text'); }
  foreach($banIP as $ip)
  {
    echo "IP $ip has been banned.\n";
    $val = array($ip, $reason, time(), $time);
    Core::getQuery()->insert("ban", $att, $val);
  }
  exit;
}

/**
* Checks length of binary number and add zeros as necessary.
*
* @param string Binary number.
* @param integer Required length.
*
* @return string Validated binary number.
*/
function validateBin($bin, $len)
{
  if($len > strlen($bin))
  {
    for($i = 0; $i < $len - strlen($bin); $i++)
    {
      $bin = "0".$bin;
    }
  }
  return $bin;
}
?>
