/**
 * Fleet related JavaScript functions.
 *
 * Oxsar http://oxsar.ru
 *
 *
 */

function selectShips()
{
	for(i = 0; i < fleet.length; i++)
	{
		var quantity = quantities[fleet[i]];
		$('#ship_'+fleet[i]).val(quantity);
		resetRest('ship_'+fleet[i]);
	}
}

function deselectShips()
{
	for(i = 0; i < fleet.length; i++)
	{
		$('#ship_'+fleet[i]).val(0);
	}
}

function getFlyTime(distance, maxspeed, speed)
{
	var time = Math.round((35000 / speed) * Math.sqrt(distance * 10 / maxspeed) + 10);
	if(gamespeed > 0)
	{
		time *= gamespeed;
	}
	return Math.ceil(time);
}

function getFlyConsumption(basicConsumption, distance, speed)
{
	return Math.ceil(basicConsumption * distance / 35000 * ((speed / 10) + 1) * ((speed / 10) + 1));
}

function getDistance(galaxy, system, pos)
{
	if(galaxy - oGalaxy != 0)
	{
		return Math.abs(galaxy - oGalaxy) * galaxy_distance_mult;
	}
	else if(system - oSystem != 0)
	{
		if( maxSystem > 0 )
		{
			var between	= Math.abs(system - oSystem);
			var around	= maxSystem - Math.max( system, oSystem ) + Math.min( system, oSystem );
			return Math.min( between, around ) * 5 * 19 + 2700;
		}
		return Math.abs(system - oSystem) * 5 * 19 + 2700;
	}
	else if(pos - oPos != 0)
	{
		return Math.abs(pos - oPos) * 5 + 1000;
	}
	return 5;
}

function rebuild()
{
	// Get vars
	var speed = $('#speed').val() / 10;
	var galaxy = $('#galaxy').val();
	var system = $('#system').val();
	var pos = $('#position').val();

	// Validate it
	if(speed < 0.1) { speed = 0.1; }
	else if(speed > 10) { speed = 10; }
	if(galaxy < 1) { galaxy = 1; }
	else if(galaxy > maxGalaxy) { galaxy = maxGalaxy; }
	if(system < 1) { system = 1; }
	else if(system > maxSystem) { system = maxSystem; }
	if(pos < 1) { pos = 1; }
	else if(pos > maxPos) { pos = maxPos; }

	var flag = false;
	if( expPlanetPos > 0 && expVirtPlanetPos > 0 && pos == expPlanetPos )
	{
		flag = true;
		pos  = expVirtPlanetPos;
	}

	// Calculations
	var distance = getDistance(galaxy, system, pos);
	var consumption = getFlyConsumption(basicConsumption, distance, speed);
	var time = getFlyTime(distance, maxspeed, speed);
	var stargate_transport_time = Math.ceil(time * stargate_transport_time_scale);
	var stargate_transport_consumption = Math.ceil(consumption * stargate_transport_consumption_scale);
	var fleetConsumption = 0;
	if( fleetSize > 0 )
	{
		var unitConsumptionPerHour = Math.min(maxGroupUnitConsumptionPerHour, (Math.pow(unitsGroupConsumptionPowerBase, fleetSize) / 10) / 24);
        var groupConsumptionPerHour = unitConsumptionPerHour * fleetSize;
		fleetConsumption = groupConsumptionPerHour >= 0.01 ? groupConsumptionPerHour * time * 2 / (60 * 60) : 0;
		if(consumption < fleetConsumption) consumption = fleetConsumption;
		if(stargate_transport_consumption < fleetConsumption) stargate_transport_consumption = fleetConsumption;
	}
	if( flag == true )
	{
		pos = expPlanetPos;
	}
	// Write it
	$('#time').text(secToString(time));
	$('#fuel').text(fNumber(consumption));
	$('#stargate_transport_time').text(secToString(stargate_transport_time));
	$('#stargate_transport_fuel').text(fNumber(stargate_transport_consumption));
	$('#distance').text(fNumber(distance));
	$('#capicity').text(fNumber(capicity - consumption));
	$('#speed').val(speed * 10);
	$('#galaxy').val(galaxy);
	$('#system').val(system);
	$('#position').val(pos);

	// Format
	if(capicity - consumption > 0)
	{
		$('#capacity').addClass('true');
		$('#fuel').addClass('true');
		$('#capacity').removeClass('false');
		$('#fuel').removeClass('false');
	}
	else
	{
		$('#capacity').removeClass('true');
		$('#fuel').removeClass('true');
		$('#capacity').addClass('false');
		$('#fuel').addClass('false');
	}
}

function rebuildLot()
{
    var consumption = parseSafeInt($('#delivery_hydro').val());
    if( consumption > capicity )
	{
    	consumption = capicity;
    	$('#delivery_hydro').val(consumption)
	}
    $('#capicity').text(fNumber(capicity - consumption));
}

function setAllResources()
{
	if(capacity < tMetal) { setMetal = capacity; }
	else { setMetal = tMetal; }
	setMaxRes('metal', setMetal);

	if(capacity < tSilicon) { setSilicon = capacity; }
	else { setSilicon = tSilicon; }
	setMaxRes('silicon', setSilicon);

	if(capacity < tHydrogen) { setHydrogen = capacity; }
	else { setHydrogen = tHydrogen; }
	setMaxRes('hydrogen', setHydrogen);
}

function setNoResources()
{
	setMinRes('metal');
	setMinRes('silicon');
	setMinRes('hydrogen');
}

function setMinRes(id)
{
	newVal = getValueFromId(id, true);
	if(id == "metal")
	{
		tMetal += newVal;
	}
	else if(id == "silicon")
	{
		tSilicon += newVal;
	}
	else
	{
		tHydrogen += newVal;
	}
	capacity += newVal;
	$('#'+id).val(0)
	setRest();
}

function setMaxRes(id, value)
{
	value = parseSafeInt(value);
	obj = document.getElementById(id);
	if(value > capacity)
	{
		value = capacity;
		capacity = 0;
	}
	else { capacity -= value; }
	if(id == 'metal')
	{
		tMetal -= value;
	}
	else if(id == 'silicon')
	{
		tSilicon -= value;
	}
	else if(id == 'hydrogen')
	{
		tHydrogen -= value;
	}
	add = getValueFromId(id, true);
	obj.value = value + add;

	setRest();
}

function renewTransportRes()
{
	var inMetal = getValueFromId('metal', true);
	var inSilicon = getValueFromId('silicon', true);
	var inHydrogen = getValueFromId('hydrogen', true);

	tMetal = outMetal - inMetal;
	tSilicon = outSilicon - inSilicon;
	tHydrogen = outHydrogen - inHydrogen;
	capacity = sCapacity - inMetal - inSilicon - inHydrogen;

	setRest();
}

function setCoordinates(galaxy, system, position, type)
{
	$('#galaxy').val(galaxy);
	$('#system').val(system);
	$('#position').val(position);
	document.getElementById('targetType').selectedIndex = type;
	rebuild();
}

function getValueFromId(id, integer)
{
	ret = $('#'+id).val();
	if(integer)
	{
		ret = parseSafeInt(ret);
		// if(isNaN(ret)) { ret = 0; }
		return ret;
	}
	return ret;
}

function setRest()
{
	obj = $('#rest');
	obj.text(fNumber(capacity));
	if(capacity < 0)
	{
		obj.addClass('false');
		obj.removeClass('true');
	}
	else
	{
		obj.addClass('true');
		obj.removeClass('false');
	}
}