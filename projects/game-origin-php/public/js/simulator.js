/**
 * Fleet related JavaScript functions.
 * 
 * Oxsar http://oxsar.ru
 *
 * 
 */

function setAllUnits(prefix, unit_type)
{
  // alert("setAllUnits: " + sim_units.length + ", " + sim_units[0].unit_type); // .length));
	for(i = 0; i < sim_units.length; i++)
	{
	  if(sim_units[i].unit_type == unit_type)
	  {
		  setUnitQuantity(prefix+'unit_'+sim_units[i].id, sim_units[i].quantity, sim_units[i].damaged, sim_units[i].shell_percent);
	  }
	}
}

function setAllUnitsShellPercent(prefix, unit_type, shell_percent)
{
	for(i = 0; i < sim_units.length; i++)
	{
	  if(sim_units[i].unit_type == unit_type)
	  {
		  setUnitQuantity(prefix+'unit_'+sim_units[i].id, 
		    -1 /*sim_units[i].quantity*/, 
		    -1 /* sim_units[i].quantity */, 
		    shell_percent);
	  }
	}
}

function resetAllUnits(unit_type)
{
	for(i = 0; i < sim_units.length; i++)
	{
	  if(sim_units[i].unit_type == unit_type)
	  {
		  setUnitQuantity('a_unit_'+sim_units[i].id, 0, 0, 100);
		  setUnitQuantity('d_unit_'+sim_units[i].id, 0, 0, 100);
	  }
	}
}

function autoIncUnitQuantity(base, num, dmg, per)
{
  if(num < 10) num = 10;
  quantity = $('#'+base).val();
  if(quantity < 1)
  {
    setUnitQuantity(base, num, dmg, per);
  }
  else if($('#'+base+'_d').val() > 0)
  {
    setUnitQuantity(base, num, 0, 100);
  }
  else
  {
    old_quantity = quantity;
    quantity *= 2;
    if((quantity > 40 && quantity < 50) || (old_quantity < 50 && quantity > 50)) quantity = 50;
    else if((quantity > 80 && quantity < 100) || (old_quantity < 100 && quantity > 100)) quantity = 100;
    else if((quantity > 400 && quantity < 500) || (old_quantity < 500 && quantity > 500)) quantity = 500;
    else if((quantity > 800 && quantity < 1000) || (old_quantity < 1000 && quantity > 1000)) quantity = 1000;
    setUnitQuantity(base, quantity, 0, 100);
  }
}

function autoDecUnitQuantity(base, num, dmg, per)
{
  quantity = $('#'+base).val();
  if($('#'+base+'_d').val() > 0)
  {
    setUnitQuantity(base, quantity, 0, 100);
  }
  else if(quantity > 0)
  {
    old_quantity = quantity;
    quantity /= 2;
    if(old_quantity > 50 && quantity < 50) quantity = 50;
    else if(old_quantity > 100 && quantity < 100) quantity = 100;
    else if(old_quantity > 500 && quantity < 500) quantity = 500;
    else if(old_quantity > 1000 && quantity < 1000) quantity = 1000;
    setUnitQuantity(base, quantity, 0, 100);
  }
}
