<?php
/**
 * This file is used for technical works. Usage:
 * @param start - start time.
 * @param end 	- end time.
 * @param type 	- Param, that can be used for titles. Used by console at the moment.
 * @param block - Does this tech work block game.
 *
 * Examples:
 * array( 'start' => '20:00:00', 'end'	=> '21:00:00', 'type' => 'hard', 'block' => true )
 * Hard block: users can't log into game and no events are executed.
 *
 * array( 'start' => '20:00:00', 'end'	=> '21:00:00', 'type' => 'light', 'block' => false )
 * Light block: users can play, but no event's are executed..
 */
return array(

	// backup
	array( 'start' => '03:30:00', 'end'	=> '03:50:00', 'block' => false, 'type' => 'light', ),
	// array( 'start' => '00:50:00', 'end'	=> '04:00:00', 'block' => true, 'type' => 'light', ),

	// single
    // array( 'start' => '14:20:00', 'end' => '15:00:00', 'block' => false, 'type' => 'light', ),

	// custom
	// array( 'start' => '00:00:00', 'end'	=> '05:00:00', 'block' => true, 'type' => 'hard' ),
);