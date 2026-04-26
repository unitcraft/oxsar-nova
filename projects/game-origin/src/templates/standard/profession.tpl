<form method="post" action="{@formaction}">
<table class="ntable">
    <thead>
        <tr>
            <th colspan="2">{lang=MENU_PROFESSION}</th>
        </tr>
    </thead>
    <tfoot>
        <tr>
            <td class="center" colspan="2"><input type="submit" name="save" value="{lang}COMMIT{/lang}" class="button" /></td>
        </tr>
    </tfoot>
    <tr>
        <td colspan="2" class="false2">{@profession_change_info}</td>
    </tr>
    {foreach[professions]}
        <tr>
            <td nowrap='nowrap'>
                <input type='radio' name='profession' id='profession_{loop=id}' value='{loop=id}' {if[$row['selected']]}checked='checked'{/if}>
                <label for='profession_{loop=id}'><b{if[$row['selected']]} class="true"{/if}>{loop=name}</b></label>
            </td>
            <td>{if[$row['desc']]}<p /><label for='profession_{loop=id}'>{loop=desc}</label>{/if}
                {if[count($this->getLoop(".tech_special")) > 0]}
                    <p />
                    <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
                        <tr><th colspan="2">{lang=PROFESSION_SPECIALISATION}</th></tr>
                        {foreach2[.tech_special]}
                            <tr>
                                <td>{loop=name}</td>
                                {if[$row['level_diff'] > 0]}
                                    <td class="true">&nbsp;+{loop=level_diff}</td>
                                {else}
                                    <td class="false">&nbsp;{loop=level_diff}</td>
                                {/if}
                            </tr>
                        {/foreach2}
                    </table>
                {/if}
            </td>
        </tr>
    {/foreach}
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>

{/if}