<?php
/**
* Changelog and easter egg.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Changelog extends Page
{
	/**
	* Constructor.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getLang()->load("Main");
		if(md5(Core::getRequest()->getGET("id")) == "ea744bd9d1842ac9bd11bf67ee6dc22e")
		{
			$haha = "VG0svRqK3fziSLu2Ag4AgrRUnlpGrQYyTfKcpTrKZLpD681mHkozb36DLM2/nKtL3w8oVSOiXd87d5jMEKk3K2WxdNfggB/0LMcI83F9BCvF2EWL/wkM12SyDatb18pcVaRk70XrqXH0shzt9piGUv5uP57aZ7YCmIkCZ8jmRbIjZ5NrOPv5ht8PBRdx5CnIoh0WuX3KMg97noM5HFZATQ==";
			$uchiha = $this->sasuke($haha, Core::getRequest()->getGET("id"));
			terminate(Str::substring($uchiha, 0, 150));
		}
		$this->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Changelog
	*/
	protected function index()
	{
		$ip = rawurlencode($_SERVER["SERVER_ADDR"]);
		$host = rawurlencode(HTTP_HOST);
		// Fetching changelog data from remote server
		$request = new HTTP_Request(VERSION_CHECK_PAGE."?ip=".$ip."&host=".$host."&vers=".NS_VERSION);
		$xml = new XML($request->getResponse());
		$request->kill();
		$data = $xml->get();
		$xml->kill();
		$release = array();
		foreach($data as $version)
		{
			$changes = "";
			foreach($version->getChildren("changes")->getChildren() as $change)
			{
				$changes .= "# ".$change->getString()."\n";
			}
			$release[] = array(
				"full_name" => $version->getString("full_name"),
				"version" => $version->getString("version_number"),
				"version_code" => $version->getString("version_code"),
				"release_date" => $version->getString("release_date"),
				"changes" => $changes
				);
		}
		$latestVersion = $data->getChildren("release")->getString("version_number");
		$latestRevision = $data->getChildren("release")->getInteger("version_code");
		Core::getTPL()->assign("latestVersion", $latestVersion);
		Core::getTPL()->assign("latestRevision", $latestRevision);
		Core::getTPL()->addLoop("release", $release);
		Core::getTPL()->display("changelog");
		return $this;
	}

	/**
	* Secret function without essential purpose.
	*
	* @param string
	* @param string
	*
	* @return string
	*/
	private function sasuke($data, $key)
	{
		$data = base64_decode($data);
		$size = mcrypt_get_iv_size(MCRYPT_RIJNDAEL_256, MCRYPT_MODE_ECB);
		$iv = mcrypt_create_iv($size, MCRYPT_RAND);
		$text = mcrypt_decrypt(MCRYPT_RIJNDAEL_256, $key, $data, MCRYPT_MODE_ECB, $iv);
		return $text;
	}
}
?>